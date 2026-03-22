package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func normalizeBaseURL(baseURL string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

func httpClientOrDefault(c *http.Client) *http.Client {
	if c != nil {
		return c
	}
	return http.DefaultClient
}

// PostJSON POSTs JSON to endpoint with Bearer token and context cancellation.
func PostJSON(client *http.Client, ctx context.Context, baseURL, endpoint string, body any, token string) (json.RawMessage, error) {
	u := normalizeBaseURL(baseURL) + "/" + strings.TrimLeft(endpoint, "/")
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	headers, err := BuildHeaders(token)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClientOrDefault(client).Do(req)
	if err != nil {
		return nil, err
	}
	return ParseJSONResponse(resp, endpoint)
}

// GetJSON performs a GET request without bot auth headers (QR endpoints).
func GetJSON(client *http.Client, ctx context.Context, baseURL, path string, extraHeaders map[string]string) (json.RawMessage, error) {
	u := normalizeBaseURL(baseURL) + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := httpClientOrDefault(client).Do(req)
	if err != nil {
		return nil, err
	}
	return ParseJSONResponse(resp, path)
}
