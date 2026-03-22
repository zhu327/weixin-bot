package weixinbot

import (
	"errors"
	"testing"
)

func TestErrMissingContextTokenWrapped(t *testing.T) {
	err := fmtMissingContextToken("alice")
	if !errors.Is(err, ErrMissingContextToken) {
		t.Fatalf("errors.Is: %v", err)
	}
}
