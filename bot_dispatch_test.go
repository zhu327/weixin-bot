package weixinbot

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSessionExpired(t *testing.T) {
	err := &APIError{Code: -14, msg: "x"}
	if !isSessionExpired(err) {
		t.Fatal()
	}
	if isSessionExpired(errors.New("other")) {
		t.Fatal()
	}
}

func TestDispatchPanicOnError(t *testing.T) {
	errCh := make(chan error, 1)
	b := New(WithOnError(func(e error) {
		select {
		case errCh <- e:
		default:
		}
	}))
	b.OnMessage(func(ctx context.Context, msg *IncomingMessage) error {
		panic("boom")
	})
	b.dispatchMessage(context.Background(), &IncomingMessage{UserID: "u"})
	select {
	case saw := <-errCh:
		if saw == nil || saw.Error() != "panic: boom" {
			t.Fatalf("got %v", saw)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for handler error")
	}
}
