package weixinbot

import "testing"

func TestExtractText(t *testing.T) {
	items := []MessageItem{
		{Type: MessageItemText, TextItem: &TextItem{Text: "hi"}},
	}
	if s := extractText(items); s != "hi" {
		t.Fatal(s)
	}
}

func TestDetectType(t *testing.T) {
	if detectType([]MessageItem{{Type: MessageItemImage}}) != KindImage {
		t.Fatal()
	}
}

func TestContextToken(t *testing.T) {
	msg := &IncomingMessage{contextToken: "tok123"}
	if msg.ContextToken() != "tok123" {
		t.Fatalf("got %q", msg.ContextToken())
	}
}

func TestContextTokenNil(t *testing.T) {
	var msg *IncomingMessage
	if msg.ContextToken() != "" {
		t.Fatal("expected empty string for nil message")
	}
}
