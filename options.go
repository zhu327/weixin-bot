package weixinbot

import "github.com/zhu327/weixin-bot/internal/api"

// Option configures [WeixinBot].
type Option func(*WeixinBot)

// WithBaseURL sets the API base URL (default: api.DefaultBaseURL).
func WithBaseURL(u string) Option {
	return func(b *WeixinBot) {
		if u != "" {
			b.baseURL = u
		}
	}
}

// WithTokenPath sets the credentials JSON path (default: ~/.weixin-bot/credentials.json).
func WithTokenPath(p string) Option {
	return func(b *WeixinBot) {
		b.tokenPath = p
	}
}

// WithOnError sets a callback for handler and transport errors (after stderr logging).
func WithOnError(fn func(error)) Option {
	return func(b *WeixinBot) {
		b.onError = fn
	}
}

// NewWeixinBot constructs a bot with optional configuration.
func NewWeixinBot(opts ...Option) *WeixinBot {
	b := &WeixinBot{
		baseURL:       api.DefaultBaseURL,
		contextTokens: make(map[string]string),
		handlers:      nil,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}
