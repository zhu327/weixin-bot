package weixinbot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDoGetUpdates_UsesHTTPClientTransport(t *testing.T) {
	var gotMethod, gotPath string
	var body map[string]any
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotMethod = req.Method
		gotPath = req.URL.Path
		b, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(b, &body)
		payload := map[string]any{
			"ret":             0,
			"msgs":            []any{},
			"get_updates_buf": "next-cursor",
		}
		raw, _ := json.Marshal(payload)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(raw)),
			Header:     make(http.Header),
		}, nil
	})
	client := &http.Client{Transport: rt}
	b := New(WithHTTPClient(client), WithBaseURL("https://example.com"))
	b.mu.Lock()
	b.credentials = &Credentials{
		Token:   "tok",
		BaseURL: "https://example.com",
	}
	b.cursor = "prev-buf"
	b.mu.Unlock()

	raw, err := b.doGetUpdates(context.Background(), b.credentials)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method: %q", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/ilink/bot/getupdates") {
		t.Fatalf("path: %q", gotPath)
	}
	if body["get_updates_buf"] != "prev-buf" {
		t.Fatalf("get_updates_buf: %v", body["get_updates_buf"])
	}
	var env struct {
		GetUpdatesBuf string `json:"get_updates_buf"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.GetUpdatesBuf != "next-cursor" {
		t.Fatalf("response buf: %q", env.GetUpdatesBuf)
	}
}
