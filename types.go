package weixinbot

import "time"

// MaxMessageTextRunes is the maximum UTF-8 code points sent per API message (long text is split).
const MaxMessageTextRunes = 2000

// Message type constants (WeChat wire values).
const (
	MessageTypeUser = 1
	MessageTypeBot  = 2
)

const (
	MessageStateNew        = 0
	MessageStateGenerating = 1
	MessageStateFinish     = 2
)

const (
	MessageItemText  = 1
	MessageItemImage = 2
	MessageItemVoice = 3
	MessageItemFile  = 4
	MessageItemVideo = 5
)

// MessageKind is the simplified inbound message kind.
type MessageKind string

const (
	KindText  MessageKind = "text"
	KindImage MessageKind = "image"
	KindVoice MessageKind = "voice"
	KindFile  MessageKind = "file"
	KindVideo MessageKind = "video"
)

// BaseInfo is included in every POST body.
type BaseInfo struct {
	ChannelVersion string `json:"channel_version"`
}

// CDNMedia is referenced by media item payloads.
type CDNMedia struct {
	EncryptQueryParam string `json:"encrypt_query_param"`
	AESKey            string `json:"aes_key"`
	EncryptType       *int   `json:"encrypt_type,omitempty"`
}

// TextItem is a text segment in item_list.
type TextItem struct {
	Text string `json:"text"`
}

// ImageItem is an image segment in item_list.
type ImageItem struct {
	Media     CDNMedia `json:"media"`
	AESKey    string   `json:"aeskey,omitempty"`
	URL       string   `json:"url,omitempty"`
	MidSize   any      `json:"mid_size,omitempty"`
	ThumbSize any      `json:"thumb_size,omitempty"`
	ThumbH    int      `json:"thumb_height,omitempty"`
	ThumbW    int      `json:"thumb_width,omitempty"`
	HDSize    any      `json:"hd_size,omitempty"`
}

// VoiceItem is a voice segment in item_list.
type VoiceItem struct {
	Media      CDNMedia `json:"media"`
	EncodeType int      `json:"encode_type,omitempty"`
	Text       string   `json:"text,omitempty"`
	Playtime   int      `json:"playtime,omitempty"`
}

// FileItem is a file segment in item_list.
type FileItem struct {
	Media    CDNMedia `json:"media"`
	FileName string   `json:"file_name,omitempty"`
	MD5      string   `json:"md5,omitempty"`
	Len      string   `json:"len,omitempty"`
}

// VideoItem is a video segment in item_list.
type VideoItem struct {
	Media      CDNMedia  `json:"media"`
	VideoSize  any       `json:"video_size,omitempty"`
	PlayLength int       `json:"play_length,omitempty"`
	ThumbMedia *CDNMedia `json:"thumb_media,omitempty"`
}

// RefMessage is optional reference metadata.
type RefMessage struct {
	Title       string       `json:"title,omitempty"`
	MessageItem *MessageItem `json:"message_item,omitempty"`
}

// MessageItem is one element of item_list.
type MessageItem struct {
	Type      int         `json:"type"`
	TextItem  *TextItem   `json:"text_item,omitempty"`
	ImageItem *ImageItem  `json:"image_item,omitempty"`
	VoiceItem *VoiceItem  `json:"voice_item,omitempty"`
	FileItem  *FileItem   `json:"file_item,omitempty"`
	VideoItem *VideoItem  `json:"video_item,omitempty"`
	RefMsg    *RefMessage `json:"ref_msg,omitempty"`
}

// WeixinMessage is a message from getupdates or sendmessage payloads.
type WeixinMessage struct {
	MessageID    int64         `json:"message_id"`
	FromUserID   string        `json:"from_user_id"`
	ToUserID     string        `json:"to_user_id"`
	ClientID     string        `json:"client_id"`
	CreateTimeMs int64         `json:"create_time_ms"`
	MessageType  int           `json:"message_type"`
	MessageState int           `json:"message_state"`
	ContextToken string        `json:"context_token"`
	ItemList     []MessageItem `json:"item_list"`
}

// IncomingMessage is the high-level view passed to handlers.
type IncomingMessage struct {
	UserID    string
	Text      string
	Type      MessageKind
	Raw       WeixinMessage
	Timestamp time.Time

	contextToken string
}
