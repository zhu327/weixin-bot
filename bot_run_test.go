package weixinbot

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunWaitsForSlowHandler(t *testing.T) {
	const oneUserMsg = `{"ret":0,"msgs":[{"message_id":1,"from_user_id":"uid1","to_user_id":"","client_id":"c","create_time_ms":0,"message_type":1,"message_state":2,"context_token":"ctxok","item_list":[{"type":1,"text_item":{"text":"hi"}}]}],"get_updates_buf":""}`

	var count int
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || !strings.Contains(req.URL.Path, "getupdates") {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		count++
		if count == 1 {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(oneUserMsg))),
				Header:     make(http.Header),
			}, nil
		}
		<-req.Context().Done()
		return nil, req.Context().Err()
	})

	b := New(
		WithHTTPClient(&http.Client{Transport: rt}),
		WithBaseURL("https://example.com"),
		WithHandlerDrainTimeout(3*time.Second),
	)
	b.mu.Lock()
	b.credentials = &Credentials{Token: "tok", BaseURL: "https://example.com"}
	b.mu.Unlock()

	started := make(chan struct{})
	var once sync.Once
	unblock := make(chan struct{})
	b.OnMessage(func(ctx context.Context, msg *IncomingMessage) error {
		once.Do(func() { close(started) })
		<-unblock
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- b.Run(ctx) }()

	<-started
	cancel()
	close(unblock)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after handler completed")
	}
}
