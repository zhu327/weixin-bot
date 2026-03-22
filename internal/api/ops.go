package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

func baseInfo() map[string]any {
	return map[string]any{"channel_version": ChannelVersion}
}

// GetUpdates calls POST /ilink/bot/getupdates.
func GetUpdates(ctx context.Context, baseURL, token, buf string) (json.RawMessage, error) {
	body := map[string]any{
		"get_updates_buf": buf,
		"base_info":       baseInfo(),
	}
	return PostJSON(ctx, baseURL, "/ilink/bot/getupdates", body, token, 40*time.Second)
}

// SendMessage calls POST /ilink/bot/sendmessage.
func SendMessage(ctx context.Context, baseURL, token string, msg map[string]any) (json.RawMessage, error) {
	body := map[string]any{
		"msg":       msg,
		"base_info": baseInfo(),
	}
	return PostJSON(ctx, baseURL, "/ilink/bot/sendmessage", body, token, 15*time.Second)
}

// GetConfig calls POST /ilink/bot/getconfig.
func GetConfig(ctx context.Context, baseURL, token, userID, contextToken string) (json.RawMessage, error) {
	body := map[string]any{
		"ilink_user_id": userID,
		"context_token": contextToken,
		"base_info":     baseInfo(),
	}
	return PostJSON(ctx, baseURL, "/ilink/bot/getconfig", body, token, 15*time.Second)
}

// SendTyping calls POST /ilink/bot/sendtyping. status 1 = start, 2 = stop.
func SendTyping(ctx context.Context, baseURL, token, userID, ticket string, status int) (json.RawMessage, error) {
	body := map[string]any{
		"ilink_user_id": userID,
		"typing_ticket": ticket,
		"status":        status,
		"base_info":     baseInfo(),
	}
	return PostJSON(ctx, baseURL, "/ilink/bot/sendtyping", body, token, 15*time.Second)
}

// FetchQRCode GET /ilink/bot/get_bot_qrcode?bot_type=3
func FetchQRCode(ctx context.Context, baseURL string) (json.RawMessage, error) {
	return GetJSON(ctx, baseURL, "/ilink/bot/get_bot_qrcode?bot_type=3", nil)
}

// PollQRStatus GET /ilink/bot/get_qrcode_status?qrcode=...
func PollQRStatus(ctx context.Context, baseURL, qrcode string) (json.RawMessage, error) {
	q := url.QueryEscape(qrcode)
	path := fmt.Sprintf("/ilink/bot/get_qrcode_status?qrcode=%s", q)
	return GetJSON(ctx, baseURL, path, map[string]string{
		"iLink-App-ClientVersion": "1",
	})
}

// RandomUUID returns a random UUID v4 string.
func RandomUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// BuildTextMessage builds the outbound text message payload (msg field body).
func BuildTextMessage(userID, contextToken, text string) (map[string]any, error) {
	clientID, err := RandomUUID()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"from_user_id":  "",
		"to_user_id":    userID,
		"client_id":     clientID,
		"message_type":  2, // BOT
		"message_state": 2, // FINISH
		"context_token": contextToken,
		"item_list": []any{
			map[string]any{
				"type": 1,
				"text_item": map[string]any{
					"text": text,
				},
			},
		},
	}, nil
}

// DecodeTypingTicket extracts typing_ticket from getconfig JSON response.
func DecodeTypingTicket(raw json.RawMessage) (string, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", false
	}
	v, ok := m["typing_ticket"]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil || s == "" {
		return "", false
	}
	return s, true
}
