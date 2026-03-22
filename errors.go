package weixinbot

import (
	"encoding/json"
	"fmt"
)

// ApiError is returned when the HTTP layer or API ret/errcode indicates failure.
type ApiError struct {
	Status  int
	Code    int
	Payload json.RawMessage
	msg     string
}

func (e *ApiError) Error() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// SessionExpired reports whether the API indicated an expired session (errcode -14).
func (e *ApiError) SessionExpired() bool {
	return e != nil && e.Code == -14
}

// Unwrap is not used; ApiError is a concrete type for errors.Is helpers if needed later.

func fmtMissingContextToken(userID string) error {
	return fmt.Errorf("no cached context token for user %s: reply to an incoming message first", userID)
}

func errEmptyMessageText() error {
	return fmt.Errorf("message text cannot be empty")
}
