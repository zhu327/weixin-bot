package weixinbot

import "testing"

func TestChunkRunes(t *testing.T) {
	s := string([]rune{0x4e00, 0x4e01}) + string(make([]rune, 1999))
	chunks := chunkRunes(s, MaxMessageTextRunes)
	if len(chunks) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(chunks))
	}
}

func TestChunkRunesEmpty(t *testing.T) {
	ch := chunkRunes("", MaxMessageTextRunes)
	if len(ch) != 1 || ch[0] != "" {
		t.Fatalf("%#v", ch)
	}
}
