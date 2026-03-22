package weixinbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/zhu327/weixin-bot/internal/api"
	"github.com/zhu327/weixin-bot/internal/auth"
)

// MessageHandler processes an inbound user message.
type MessageHandler func(ctx context.Context, msg *IncomingMessage) error

// LoginOptions controls [Bot.Login].
type LoginOptions struct {
	Force bool
}

// Bot is the main SDK entrypoint (Node WeixinBot).
type Bot struct {
	baseURL    string
	tokenPath  string
	onError    func(error)
	httpClient *http.Client
	logger     *slog.Logger

	handlers []MessageHandler

	mu          sync.Mutex
	tokenCache  *contextTokenCache
	credentials *Credentials
	cursor      string

	runMu     sync.Mutex
	activeRun *runSession
}

// WeixinBot is a type alias for [Bot].
//
// Deprecated: use [Bot] instead.
type WeixinBot = Bot

type runSession struct {
	done chan struct{}
	err  error
}

// OnMessage registers a handler (multiple handlers run concurrently per message).
func (b *Bot) OnMessage(h MessageHandler) *Bot {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, h)
	return b
}

// Login performs QR or loads stored credentials (see spec).
func (b *Bot) Login(ctx context.Context, opts LoginOptions) (*Credentials, error) {
	path, err := auth.ResolveTokenPath(b.tokenPath)
	if err != nil {
		return nil, err
	}
	prevTok := ""
	b.mu.Lock()
	if b.credentials != nil {
		prevTok = b.credentials.Token
	}
	b.mu.Unlock()

	d, err := auth.Login(ctx, b.httpClient, b.baseURL, path, opts.Force)
	if err != nil {
		return nil, err
	}
	cred := credentialsFromData(d)
	b.mu.Lock()
	b.credentials = cred
	b.baseURL = cred.BaseURL
	if prevTok != "" && prevTok != cred.Token {
		b.clearCursorAndContextTokens()
	}
	b.mu.Unlock()

	b.logf("Logged in as %s\n", cred.UserID)
	return cred, nil
}

// Run starts the long-poll loop until ctx is canceled.
func (b *Bot) Run(ctx context.Context) error {
	b.runMu.Lock()
	if b.activeRun != nil {
		s := b.activeRun
		b.runMu.Unlock()
		<-s.done
		return s.err
	}
	s := &runSession{done: make(chan struct{})}
	b.activeRun = s
	b.runMu.Unlock()

	err := b.runLoop(ctx)

	b.runMu.Lock()
	s.err = err
	close(s.done)
	b.activeRun = nil
	b.runMu.Unlock()
	return err
}

func (b *Bot) runLoop(ctx context.Context) error {
	if _, err := b.ensureCredentials(ctx); err != nil {
		return err
	}
	b.logf("Long-poll loop started.\n")
	retry := time.Second

	for {
		select {
		case <-ctx.Done():
			b.logf("Long-poll loop stopped.\n")
			return nil
		default:
		}

		cred, err := b.ensureCredentials(ctx)
		if err != nil {
			b.reportError(err)
			if err := b.sleepBackoff(ctx, retry); err != nil {
				b.logf("Long-poll loop stopped.\n")
				return err
			}
			retry = min(retry*2, 10*time.Second)
			continue
		}

		raw, err := b.doGetUpdates(ctx, cred)
		if err != nil {
			if ctx.Err() != nil {
				b.logf("Long-poll loop stopped.\n")
				return nil
			}
			if isSessionExpired(err) {
				b.logf("Session expired. Waiting for a fresh QR login...\n")
				b.mu.Lock()
				b.credentials = nil
				b.clearCursorAndContextTokens()
				b.mu.Unlock()
				path, _ := auth.ResolveTokenPath(b.tokenPath)
				_ = auth.Clear(path)
				if _, lerr := b.Login(ctx, LoginOptions{Force: true}); lerr != nil {
					b.reportError(lerr)
				}
				retry = time.Second
				continue
			}
			b.reportError(err)
			if err := b.sleepBackoff(ctx, retry); err != nil {
				b.logf("Long-poll loop stopped.\n")
				return err
			}
			retry = min(retry*2, 10*time.Second)
			continue
		}

		retry = time.Second
		var env struct {
			Msgs          []json.RawMessage `json:"msgs"`
			GetUpdatesBuf string            `json:"get_updates_buf"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			b.reportError(err)
			continue
		}
		if env.GetUpdatesBuf != "" {
			b.mu.Lock()
			b.cursor = env.GetUpdatesBuf
			b.mu.Unlock()
		}
		for _, mr := range env.Msgs {
			wm, err := decodeWeixinMessage(mr)
			if err != nil {
				b.reportError(err)
				continue
			}
			b.rememberContext(wm)
			inc := toIncomingMessage(wm)
			if inc == nil {
				continue
			}
			b.dispatchMessage(ctx, inc)
		}
	}
}

func (b *Bot) doGetUpdates(ctx context.Context, cred *Credentials) (json.RawMessage, error) {
	b.mu.Lock()
	buf := b.cursor
	b.mu.Unlock()
	base := cred.BaseURL
	raw, err := api.GetUpdates(b.httpClient, ctx, base, cred.Token, buf)
	return raw, wrapAPIErr(err)
}

func isSessionExpired(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.SessionExpired()
}

func (b *Bot) sleepBackoff(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (b *Bot) ensureCredentials(ctx context.Context) (*Credentials, error) {
	b.mu.Lock()
	if b.credentials != nil {
		c := b.credentials
		b.mu.Unlock()
		return c, nil
	}
	b.mu.Unlock()

	path, err := auth.ResolveTokenPath(b.tokenPath)
	if err != nil {
		return nil, err
	}
	d, err := auth.Load(path)
	if err != nil {
		return nil, err
	}
	if d != nil {
		c := credentialsFromData(d)
		b.mu.Lock()
		b.credentials = c
		b.baseURL = c.BaseURL
		b.mu.Unlock()
		return c, nil
	}
	_, err = b.Login(ctx, LoginOptions{Force: false})
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.credentials, nil
}

func (b *Bot) dispatchMessage(ctx context.Context, msg *IncomingMessage) {
	b.mu.Lock()
	hcopy := append([]MessageHandler(nil), b.handlers...)
	b.mu.Unlock()
	if len(hcopy) == 0 {
		return
	}
	for _, h := range hcopy {
		h := h
		go func() {
			defer func() {
				if r := recover(); r != nil {
					b.reportError(fmt.Errorf("panic: %v", r))
				}
			}()
			if err := h(ctx, msg); err != nil {
				b.reportError(err)
			}
		}()
	}
}

// Reply sends a text reply and clears typing (errors from StopTyping ignored).
func (b *Bot) Reply(ctx context.Context, msg *IncomingMessage, text string) error {
	if msg == nil {
		return fmt.Errorf("%w", ErrNilMessage)
	}
	b.tokenCache.Set(msg.UserID, msg.contextToken)
	if err := b.sendText(ctx, msg.UserID, text, msg.contextToken); err != nil {
		return err
	}
	_ = b.StopTyping(ctx, msg.UserID)
	return nil
}

// Send sends text using the cached context token for userID.
func (b *Bot) Send(ctx context.Context, userID, text string) error {
	tok, ok := b.tokenCache.Get(userID)
	if !ok || tok == "" {
		return fmtMissingContextToken(userID)
	}
	return b.sendText(ctx, userID, text, tok)
}

func (b *Bot) sendText(ctx context.Context, userID, text, contextToken string) error {
	if text == "" {
		return fmt.Errorf("%w", ErrEmptyMessageText)
	}
	cred, err := b.ensureCredentials(ctx)
	if err != nil {
		return err
	}
	for _, chunk := range chunkRunes(text, MaxMessageTextRunes) {
		msg, err := api.BuildTextMessage(userID, contextToken, chunk)
		if err != nil {
			return err
		}
		if _, err := api.SendMessage(b.httpClient, ctx, cred.BaseURL, cred.Token, msg); err != nil {
			return wrapAPIErr(err)
		}
	}
	return nil
}

// SendTyping requests a typing indicator for userID.
func (b *Bot) SendTyping(ctx context.Context, userID string) error {
	return b.sendTypingStatus(ctx, userID, api.TypingStatusStart, true)
}

// StopTyping clears the typing indicator (no-op if no cached context).
func (b *Bot) StopTyping(ctx context.Context, userID string) error {
	return b.sendTypingStatus(ctx, userID, api.TypingStatusStop, false)
}

// sendTypingStatus calls sendtyping with status 1=start, 2=stop. logIfNoTicket applies when status is start.
func (b *Bot) sendTypingStatus(ctx context.Context, userID string, status int, logIfNoTicket bool) error {
	ct, ok := b.tokenCache.Get(userID)
	if !ok || ct == "" {
		if status == api.TypingStatusStart {
			return fmtMissingContextToken(userID)
		}
		return nil
	}
	cred, err := b.ensureCredentials(ctx)
	if err != nil {
		return err
	}
	raw, err := api.GetConfig(b.httpClient, ctx, cred.BaseURL, cred.Token, userID, ct)
	if err != nil {
		return wrapAPIErr(err)
	}
	ticket, ok := api.DecodeTypingTicket(raw)
	if !ok {
		if logIfNoTicket {
			b.logf("sendTyping: no typing_ticket returned by getconfig\n")
		}
		return nil
	}
	_, err = api.SendTyping(b.httpClient, ctx, cred.BaseURL, cred.Token, userID, ticket, status)
	return wrapAPIErr(err)
}

// clearCursorAndContextTokens resets long-poll cursor and per-user context cache. b.mu must be held.
func (b *Bot) clearCursorAndContextTokens() {
	b.cursor = ""
	b.tokenCache.Clear()
}

func wrapAPIErr(err error) error {
	if err == nil {
		return nil
	}
	var ae *api.APIError
	if errors.As(err, &ae) {
		return &APIError{Status: ae.Status, Code: ae.Code, Payload: ae.Payload, msg: ae.Error()}
	}
	return err
}
