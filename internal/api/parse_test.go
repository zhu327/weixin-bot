package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestParseJSONResponse_HTTPError(t *testing.T) {
	body := `{"errmsg":"bad","errcode":401}`
	resp := &http.Response{
		StatusCode: 401,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	_, err := ParseJSONResponse(resp, "/test")
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want *APIError, got %T", err)
	}
	if ae.Status != 401 {
		t.Errorf("status: %d", ae.Status)
	}
}

func TestParseJSONResponse_RetNonZero(t *testing.T) {
	body := `{"ret":1,"errmsg":"fail","errcode":-14}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	_, err := ParseJSONResponse(resp, "/test")
	if err == nil {
		t.Fatal("expected error")
	}
	ae := err.(*APIError)
	if ae.Code != -14 && ae.Code != 1 {
		t.Errorf("code: %d", ae.Code)
	}
}

func TestParseJSONResponse_OK(t *testing.T) {
	body := `{"ret":0,"msgs":[]}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	raw, err := ParseJSONResponse(resp, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("msgs")) {
		t.Fatalf("unexpected body: %s", raw)
	}
}

func TestParseJSONResponse_OKNoRet(t *testing.T) {
	body := `{"foo":1}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
	_, err := ParseJSONResponse(resp, "/test")
	if err != nil {
		t.Fatal(err)
	}
}
