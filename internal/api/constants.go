package api

import "time"

// DefaultBaseURL is the public iLink endpoint (Node DEFAULT_BASE_URL).
const DefaultBaseURL = "https://ilinkai.weixin.qq.com"

// ChannelVersion is sent in base_info on every POST.
const ChannelVersion = "1.0.0"

// HTTP timeouts for outbound requests (used with context.WithTimeout).
const (
	TimeoutGetUpdates   = 40 * time.Second
	TimeoutSendMessage  = 15 * time.Second
	TimeoutGetConfig    = 15 * time.Second
	TimeoutSendTyping   = 15 * time.Second
	TimeoutFetchQRCode  = 30 * time.Second
	TimeoutPollQRStatus = 30 * time.Second
)

// BotQRType is the bot_type query value for get_bot_qrcode.
const BotQRType = 3

// Outbound wire values for BuildTextMessage (message_type / message_state).
const (
	OutboundMessageTypeBot         = 2
	OutboundMessageStateGenerating = 1
	OutboundMessageStateFinish     = 2
	OutboundItemTypeText           = 1
)

// TypingStatusStart and TypingStatusStop are sendtyping status values.
const (
	TypingStatusStart = 1
	TypingStatusStop  = 2
)
