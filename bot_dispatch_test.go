package weixinbot

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestSessionExpired(t *testing.T) {
	err := &ApiError{Code: -14, msg: "x"}
	if !isSessionExpired(err) {
		t.Fatal()
	}
	if isSessionExpired(errors.New("other")) {
		t.Fatal()
	}
}

func TestDispatchPanicOnError(t *testing.T) {
	var saw error
	var mu sync.Mutex
	b := NewWeixinBot(WithOnError(func(e error) {
		mu.Lock()
		saw = e
		mu.Unlock()
	}))
	b.OnMessage(func(ctx context.Context, msg *IncomingMessage) error {
		panic("boom")
	})
	b.dispatchMessage(context.Background(), &IncomingMessage{UserID: "u"})
	mu.Lock()
	defer mu.Unlock()
	if saw == nil || saw.Error() != "panic: boom" {
		t.Fatalf("got %v", saw)
	}
}
