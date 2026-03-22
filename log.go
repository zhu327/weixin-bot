package weixinbot

import (
	"fmt"
	"os"
)

func (b *WeixinBot) logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[weixin-bot] "+format, args...)
}

func (b *WeixinBot) reportError(err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	b.logf("%s\n", msg)
	if b.onError != nil {
		b.onError(err)
	}
}
