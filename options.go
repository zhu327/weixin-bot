package weixinbot

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/zhu327/weixin-bot/internal/api"
)

// Option configures [Bot].
type Option func(*Bot)

// WithBaseURL sets the API base URL (default: api.DefaultBaseURL).
func WithBaseURL(u string) Option {
	return func(b *Bot) {
		if u != "" {
			b.baseURL = u
		}
	}
}

// WithTokenPath sets the credentials JSON path (default: first writable of ~/.weixin-bot, ./.weixin-bot, $TMP/weixin-bot).
func WithTokenPath(p string) Option {
	return func(b *Bot) {
		b.tokenPath = p
	}
}

// WithOnError sets a callback for handler and transport errors (after logging).
func WithOnError(fn func(error)) Option {
	return func(b *Bot) {
		b.onError = fn
	}
}

// WithHTTPClient sets the HTTP client used for all API and long-poll requests (including QR login).
// If unset, [http.DefaultClient] is used by the low-level transport helpers when the field is empty;
// [New] installs a non-nil *http.Client with no per-client timeout (deadlines come from context).
func WithHTTPClient(c *http.Client) Option {
	return func(b *Bot) {
		if c != nil {
			b.httpClient = c
		}
	}
}

// WithLogger sets the slog logger for SDK diagnostics. If unset, logs are discarded.
func WithLogger(l *slog.Logger) Option {
	return func(b *Bot) {
		if l != nil {
			b.logger = l
		}
	}
}

// WithContextTokenCacheMax sets the LRU capacity for per-user context_token entries (default 10000).
func WithContextTokenCacheMax(n int) Option {
	return func(b *Bot) {
		if n > 0 {
			b.tokenCache = newContextTokenCache(n)
		}
	}
}

// WithHandlerDrainTimeout sets how long [Bot.Run] waits for in-flight message handlers after the
// long-poll loop stops. Zero selects the default (30s). A negative value skips waiting (handlers may
// still be running when Run returns).
func WithHandlerDrainTimeout(d time.Duration) Option {
	return func(b *Bot) {
		b.handlerDrainWait = d
	}
}

// New constructs a bot with optional configuration.
func New(opts ...Option) *Bot {
	b := &Bot{
		baseURL:    api.DefaultBaseURL,
		tokenCache: newContextTokenCache(defaultContextTokenCacheMax),
		httpClient: &http.Client{},
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// NewWeixinBot constructs a bot; it is equivalent to [New].
//
// Deprecated: use [New] instead.
func NewWeixinBot(opts ...Option) *Bot {
	return New(opts...)
}
