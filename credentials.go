package weixinbot

// Credentials holds persisted bot login data (same JSON shape as Node auth).
type Credentials struct {
	Token     string `json:"token"`
	BaseURL   string `json:"baseUrl"`
	AccountID string `json:"accountId"`
	UserID    string `json:"userId"`
}
