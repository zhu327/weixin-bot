package weixinbot

import "testing"

func TestContextTokenCacheEvictsLRU(t *testing.T) {
	c := newContextTokenCache(2)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Set("c", "3") // evicts "a" (oldest)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected a evicted")
	}
	if v, ok := c.Get("b"); !ok || v != "2" {
		t.Fatalf("b: %v %v", v, ok)
	}
	if v, ok := c.Get("c"); !ok || v != "3" {
		t.Fatalf("c: %v %v", v, ok)
	}
	c.Get("b")      // refresh b
	c.Set("d", "4") // should evict c not b
	if _, ok := c.Get("c"); ok {
		t.Fatal("expected c evicted")
	}
	if v, ok := c.Get("b"); !ok || v != "2" {
		t.Fatal("b should remain")
	}
	if v, ok := c.Get("d"); !ok || v != "4" {
		t.Fatal("d missing")
	}
}
