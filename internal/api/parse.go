package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// retPayload is the minimal shape for success checks.
type retPayload struct {
	Ret     *int   `json:"ret"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// parseJSONResponse parses a successful JSON body or returns [*APIError].
// The exported SDK maps these errors to its own APIError type.

// APIError is the internal API error (weixinbot converts to its exported type).
type APIError struct {
	Status  int
	Code    int
	Payload json.RawMessage
	Msg     string
}

func (e *APIError) Error() string { return e.Msg }

// ParseJSONResponse reads the response body and applies Node parseJsonResponse rules.
func ParseJSONResponse(resp *http.Response, label string) (json.RawMessage, error) {
	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		if err == nil {
			err = closeErr
		}
	}
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(body)
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, &APIError{
			Msg:     fmt.Sprintf("%s: invalid JSON: %v", label, err),
			Status:  resp.StatusCode,
			Code:    0,
			Payload: raw,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := extractErrMsg(generic)
		code := extractErrCode(generic)
		if errMsg == "" {
			errMsg = fmt.Sprintf("%s failed with HTTP %d", label, resp.StatusCode)
		}
		return nil, &APIError{Msg: errMsg, Status: resp.StatusCode, Code: code, Payload: raw}
	}

	var rp retPayload
	if err := json.Unmarshal(raw, &rp); err == nil && rp.Ret != nil && *rp.Ret != 0 {
		msg := rp.ErrMsg
		if msg == "" {
			msg = fmt.Sprintf("%s failed", label)
		}
		code := rp.ErrCode
		if code == 0 {
			code = *rp.Ret
		}
		return nil, &APIError{Msg: msg, Status: resp.StatusCode, Code: code, Payload: raw}
	}

	return raw, nil
}

func extractErrMsg(m map[string]json.RawMessage) string {
	if v, ok := m["errmsg"]; ok {
		var s string
		_ = json.Unmarshal(v, &s)
		return s
	}
	return ""
}

func extractErrCode(m map[string]json.RawMessage) int {
	if v, ok := m["errcode"]; ok {
		var c int
		_ = json.Unmarshal(v, &c)
		return c
	}
	return 0
}
