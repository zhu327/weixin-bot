package weixinbot

import (
	"encoding/json"
	"strings"
	"time"
)

func (b *Bot) rememberContext(msg *WeixinMessage) {
	var userID string
	if msg.MessageType == MessageTypeUser {
		userID = msg.FromUserID
	} else {
		userID = msg.ToUserID
	}
	if userID != "" && msg.ContextToken != "" {
		b.tokenCache.Set(userID, msg.ContextToken)
	}
}

func toIncomingMessage(msg *WeixinMessage) *IncomingMessage {
	if msg == nil || msg.MessageType != MessageTypeUser {
		return nil
	}
	ms := msg.CreateTimeMs
	if ms == 0 {
		ms = time.Now().UnixMilli()
	}
	ts := time.UnixMilli(ms)
	return &IncomingMessage{
		UserID:       msg.FromUserID,
		Text:         extractText(msg.ItemList),
		Type:         detectType(msg.ItemList),
		Raw:          *msg,
		Timestamp:    ts,
		contextToken: msg.ContextToken,
	}
}

func detectType(items []MessageItem) MessageKind {
	if len(items) == 0 {
		return KindText
	}
	switch items[0].Type {
	case MessageItemImage:
		return KindImage
	case MessageItemVoice:
		return KindVoice
	case MessageItemFile:
		return KindFile
	case MessageItemVideo:
		return KindVideo
	default:
		return KindText
	}
}

func extractText(items []MessageItem) string {
	var parts []string
	for _, item := range items {
		var piece string
		switch item.Type {
		case MessageItemText:
			if item.TextItem != nil {
				piece = item.TextItem.Text
			}
		case MessageItemImage:
			if item.ImageItem != nil && item.ImageItem.URL != "" {
				piece = item.ImageItem.URL
			} else {
				piece = "[image]"
			}
		case MessageItemVoice:
			if item.VoiceItem != nil && item.VoiceItem.Text != "" {
				piece = item.VoiceItem.Text
			} else {
				piece = "[voice]"
			}
		case MessageItemFile:
			if item.FileItem != nil && item.FileItem.FileName != "" {
				piece = item.FileItem.FileName
			} else {
				piece = "[file]"
			}
		case MessageItemVideo:
			piece = "[video]"
		default:
			piece = ""
		}
		if piece != "" {
			parts = append(parts, piece)
		}
	}
	return strings.Join(parts, "\n")
}

// ContextToken returns the context_token associated with this incoming message.
// Use this when you need low-level control over message sending, e.g. for streaming
// (GENERATING → FINISH sequences). For normal replies, use [Bot.Reply] instead.
func (m *IncomingMessage) ContextToken() string {
	if m == nil {
		return ""
	}
	return m.contextToken
}

// decodeWeixinMessage unmarshals one message JSON object.
func decodeWeixinMessage(raw json.RawMessage) (*WeixinMessage, error) {
	var m WeixinMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
