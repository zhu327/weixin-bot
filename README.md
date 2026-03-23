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

## Typing 状态（"对方正在输入"）

收到消息后、回复之前，调用 `SendTyping` 让微信端显示"对方正在输入"提示。`Reply` 会自动取消 typing，无需手动调用 `StopTyping`。

```go
bot.OnMessage(func(ctx context.Context, msg *weixinbot.IncomingMessage) error {
	_ = bot.SendTyping(ctx, msg.UserID)

	// 耗时操作（调用 LLM、查数据库等）
	answer := callLLM(msg.Text)

	return bot.Reply(ctx, msg, answer) // 自动 StopTyping
})
```

> 协议实测结论：`sendtyping` 是触发"对方正在输入"的唯一可靠方式；`GENERATING` message_state 在普通 bot 会话中**不会**触发该提示。

## 流式发送（Streaming）

对于 LLM 等场景，可以使用 `SendRawMessage` 实现两种流式效果：

### 方式 A：同一 client_id + GENERATING → FINISH

使用同一个 `client_id`，先发送 `GENERATING` 状态的中间结果，最后发送 `FINISH` 完成。协议层面 API 返回 200，但实测微信客户端**仅显示第一条 GENERATING 的内容**，后续更新不渲染。仅供参考。

```go
bot.OnMessage(func(ctx context.Context, msg *weixinbot.IncomingMessage) error {
	cid, _ := weixinbot.GenerateClientID()
	ct := msg.ContextToken()

	_ = bot.SendRawMessage(ctx, msg.UserID, "思考中...", ct, cid, weixinbot.MessageStateGenerating)
	time.Sleep(2 * time.Second)

	return bot.SendRawMessage(ctx, msg.UserID, "最终回答", ct, cid, weixinbot.MessageStateFinish)
})
```

### 方式 B：不同 client_id + 每条独立 FINISH（推荐）

每条分段使用新的 `client_id`，全部以 `FINISH` 状态发送，微信端会独立显示每条消息。配合 `SendTyping` 使用效果最佳。

```go
bot.OnMessage(func(ctx context.Context, msg *weixinbot.IncomingMessage) error {
	_ = bot.SendTyping(ctx, msg.UserID)
	ct := msg.ContextToken()

	for i, chunk := range streamFromLLM(msg.Text) {
		cid, _ := weixinbot.GenerateClientID()
		state := weixinbot.MessageStateFinish
		if err := bot.SendRawMessage(ctx, msg.UserID, chunk, ct, cid, state); err != nil {
			return err
		}
	}

	_ = bot.StopTyping(ctx, msg.UserID)
	return nil
})
```

## API notes

- **Credentials path:** defaults to `credentials.json` under the first writable directory in order: `~/.weixin-bot`, `./.weixin-bot` (current working directory), then `$TMP/weixin-bot`. Cwd is tried before the temp dir so tokens are less likely to disappear with `/tmp` cleanup when home is unavailable.
- **After `Run` returns:** the SDK waits (by default up to 30s) for message handlers that are still running, so short graceful work can finish. Use `weixinbot.WithHandlerDrainTimeout` to change the limit, or a negative duration to skip waiting.
- **QR login UX:** set `LoginOptions.OnQRCode` / `OnStatus` to receive the QR link and status strings instead of printing to stderr (useful in Kubernetes or headless environments).
- Protocol details: see `PROTOCOL.md` in the upstream [weixin-bot](https://github.com/pinixai/weixin-bot) monorepo if you need field-level API documentation.

## License

MIT
