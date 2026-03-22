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
