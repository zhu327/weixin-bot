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

func TestGenerateClientID(t *testing.T) {
	id1, err := GenerateClientID()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := GenerateClientID()
	if err != nil {
		t.Fatal(err)
	}
	if id1 == "" || id2 == "" {
		t.Fatal("empty client ID")
	}
	if id1 == id2 {
		t.Fatal("expected unique client IDs")
	}
}

func TestSendRawMessage_GeneratingAndFinish(t *testing.T) {
	var calls []map[string]any
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		calls = append(calls, body)
		raw, _ := json.Marshal(map[string]any{"ret": 0})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(raw)),
			Header:     make(http.Header),
		}, nil
	})
	client := &http.Client{Transport: rt}
	b := New(WithHTTPClient(client), WithBaseURL("https://example.com"))
	b.mu.Lock()
	b.credentials = &Credentials{Token: "tok", BaseURL: "https://example.com"}
	b.mu.Unlock()

	ctx := context.Background()
	cid, _ := GenerateClientID()

	if err := b.SendRawMessage(ctx, "user1", "thinking...", "ctx-tok", cid, MessageStateGenerating); err != nil {
		t.Fatal(err)
	}
	if err := b.SendRawMessage(ctx, "user1", "done!", "ctx-tok", cid, MessageStateFinish); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}

	msg0 := calls[0]["msg"].(map[string]any)
	msg1 := calls[1]["msg"].(map[string]any)
	if msg0["client_id"] != cid || msg1["client_id"] != cid {
		t.Fatal("client_id should be reused across streaming calls")
	}
	if int(msg0["message_state"].(float64)) != MessageStateGenerating {
		t.Fatalf("first call should be GENERATING, got %v", msg0["message_state"])
	}
	if int(msg1["message_state"].(float64)) != MessageStateFinish {
		t.Fatalf("second call should be FINISH, got %v", msg1["message_state"])
	}
}

func TestSendRawMessage_FallbackContextToken(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		raw, _ := json.Marshal(map[string]any{"ret": 0})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(raw)),
			Header:     make(http.Header),
		}, nil
	})
	b := New(WithHTTPClient(&http.Client{Transport: rt}), WithBaseURL("https://example.com"))
	b.mu.Lock()
	b.credentials = &Credentials{Token: "tok", BaseURL: "https://example.com"}
	b.mu.Unlock()
	b.tokenCache.Set("user1", "cached-ctx-tok")

	err := b.SendRawMessage(context.Background(), "user1", "hi", "", "cid", MessageStateFinish)
	if err != nil {
		t.Fatal(err)
	}

	err = b.SendRawMessage(context.Background(), "unknown-user", "hi", "", "cid", MessageStateFinish)
	if err == nil {
		t.Fatal("expected error for unknown user with no context token")
	}
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
