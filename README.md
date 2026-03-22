# weixin-bot (Go)

Go SDK for the WeChat iLink Bot API: QR login, credential persistence, long-poll messaging, and replies with automatic `context_token` handling.

**Module:** [`github.com/zhu327/weixin-bot`](https://github.com/zhu327/weixin-bot)  
**Import:** `github.com/zhu327/weixin-bot`（根目录包名 `weixinbot`，代码里仍写 `weixinbot.NewWeixinBot()`）

## Repository layout

This directory is intended to be the **root of the Go module** when published (e.g. copy these files into `github.com/zhu327/weixin-bot` as the repo root, or use a subtree split). Layout:

```
.
├── go.mod
├── README.md
├── *.go                 # package weixinbot（对外 API）
├── internal/            # api, auth — 仅本模块可引用
└── cmd/
    └── echo/            # 示例机器人
```

## Requirements

- Go 1.22+

## Install

```bash
go get github.com/zhu327/weixin-bot
```

## Quick start

```go
import "github.com/zhu327/weixin-bot"

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

bot := weixinbot.NewWeixinBot()
if _, err := bot.Login(ctx, weixinbot.LoginOptions{}); err != nil {
	log.Fatal(err)
}
bot.OnMessage(func(ctx context.Context, msg *weixinbot.IncomingMessage) error {
	return bot.Reply(ctx, msg, "echo: "+msg.Text)
})
if err := bot.Run(ctx); err != nil && err != context.Canceled {
	log.Fatal(err)
}
```

运行示例：

```bash
go run ./cmd/echo
```

## API notes

- Credentials default to `~/.weixin-bot/credentials.json` (same family of SDKs as Node/Python).
- Protocol details: see `PROTOCOL.md` in the upstream [weixin-bot](https://github.com/pinixai/weixin-bot) monorepo if you need field-level API documentation.

## License

MIT
