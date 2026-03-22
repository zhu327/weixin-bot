package weixinbot

import "fmt"

func (b *Bot) logf(format string, args ...any) {
	b.logger.Info(fmt.Sprintf(format, args...))
}

func (b *Bot) reportError(err error) {
	if err == nil {
		return
	}
	b.logger.Error("weixinbot", "err", err)
	if b.onError != nil {
		b.onError(err)
	}
}
