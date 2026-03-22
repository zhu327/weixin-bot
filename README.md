# weixin-bot (Go)

Go SDK for the WeChat iLink Bot API: QR login, credential persistence, long-poll messaging, and replies with automatic `context_token` handling.

**Module:** [`github.com/zhu327/weixin-bot`](https://github.com/zhu327/weixin-bot)  
**Import:** `github.com/zhu327/weixin-bot`（根目录包名 `weixinbot`）。推荐使用 `weixinbot.New()` 与 `*weixinbot.Bot`；`NewWeixinBot` / `WeixinBot` / `ApiError` 仍可作为已弃用别名使用。

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

bot := weixinbot.New()
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

- **Credentials path:** defaults to `credentials.json` under the first writable directory in order: `~/.weixin-bot`, `./.weixin-bot` (current working directory), then `$TMP/weixin-bot`. Cwd is tried before the temp dir so tokens are less likely to disappear with `/tmp` cleanup when home is unavailable.
- **After `Run` returns:** the SDK waits (by default up to 30s) for message handlers that are still running, so short graceful work can finish. Use `weixinbot.WithHandlerDrainTimeout` to change the limit, or a negative duration to skip waiting.
- **QR login UX:** set `LoginOptions.OnQRCode` / `OnStatus` to receive the QR link and status strings instead of printing to stderr (useful in Kubernetes or headless environments).
- Protocol details: see `PROTOCOL.md` in the upstream [weixin-bot](https://github.com/pinixai/weixin-bot) monorepo if you need field-level API documentation.

## License

MIT
