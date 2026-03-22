package weixinbot

import "log/slog"

func (b *Bot) reportError(err error) {
	if err == nil {
		return
	}
	b.logger.Error("weixinbot", slog.Any("err", err))
	if b.onError != nil {
		b.onError(err)
	}
}
