package weixinbot

import "unicode/utf8"

// chunkRunes splits s into segments of at most limit UTF-8 code points (Node Array.from semantics).
func chunkRunes(s string, limit int) []string {
	if limit <= 0 {
		return []string{s}
	}
	var out []string
	for len(s) > 0 {
		n := 0
		runes := 0
		for runes < limit && n < len(s) {
			_, sz := utf8.DecodeRuneInString(s[n:])
			if sz == 0 {
				break
			}
			n += sz
			runes++
		}
		out = append(out, s[:n])
		s = s[n:]
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}
