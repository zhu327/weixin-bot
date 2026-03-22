package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhu327/weixin-bot"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot := weixinbot.NewWeixinBot()
	if _, err := bot.Login(ctx, weixinbot.LoginOptions{}); err != nil {
		log.Fatal(err)
	}

	bot.OnMessage(func(ctx context.Context, msg *weixinbot.IncomingMessage) error {
		log.Printf("[%s] %s: %s", msg.Timestamp.In(time.Local).Format(time.Kitchen), msg.UserID, msg.Text)
		return bot.Reply(ctx, msg, "你说了: "+msg.Text)
	})

	log.Println("Bot is running. Press Ctrl+C to stop.")
	if err := bot.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
