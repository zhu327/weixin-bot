package weixinbot

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Sentinel errors for programmatic handling with [errors.Is].
var (
	ErrMissingContextToken = errors.New("weixinbot: missing context token")
	ErrEmptyMessageText    = errors.New("weixinbot: message text cannot be empty")
	ErrNilMessage          = errors.New("weixinbot: nil message")
)

// APIError is returned when the HTTP layer or API ret/errcode indicates failure.
type APIError struct {
	Status  int
	Code    int
	Payload json.RawMessage
	msg     string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// SessionExpired reports whether the API indicated an expired session (errcode -14).
func (e *APIError) SessionExpired() bool {
	return e != nil && e.Code == -14
}

// ApiError is a type alias for [APIError].
//
// Deprecated: use [APIError] instead.
type ApiError = APIError

func fmtMissingContextToken(userID string) error {
	return fmt.Errorf("%w for user %s: reply to an incoming message first", ErrMissingContextToken, userID)
}
