package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Data is persisted credential content (language-agnostic).
type Data struct {
	Token     string
	BaseURL   string
	AccountID string
	UserID    string
}

type fileFormat struct {
	Token      string `json:"token"`
	BaseURL    string `json:"baseUrl"`
	BaseURL2   string `json:"base_url"`
	AccountID  string `json:"accountId"`
	AccountID2 string `json:"account_id"`
	UserID     string `json:"userId"`
	UserID2    string `json:"user_id"`
}

// Save writes credentials with directory 0700 and file 0600 (Node parity).
func Save(path string, d *Data) error {
	if d == nil {
		return errors.New("nil credentials")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	payload := map[string]string{
		"token":     d.Token,
		"baseUrl":   d.BaseURL,
		"accountId": d.AccountID,
		"userId":    d.UserID,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// Load reads credentials; returns (nil, nil) if the file is missing.
func Load(path string) (*Data, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ff fileFormat
	if err := json.Unmarshal(b, &ff); err != nil {
		return nil, fmt.Errorf("invalid credentials format in %s", path)
	}
	baseURL := ff.BaseURL
	if baseURL == "" {
		baseURL = ff.BaseURL2
	}
	accountID := ff.AccountID
	if accountID == "" {
		accountID = ff.AccountID2
	}
	userID := ff.UserID
	if userID == "" {
		userID = ff.UserID2
	}
	if ff.Token == "" || baseURL == "" || accountID == "" || userID == "" {
		return nil, fmt.Errorf("invalid credentials format in %s", path)
	}
	return &Data{
		Token:     ff.Token,
		BaseURL:   baseURL,
		AccountID: accountID,
		UserID:    userID,
	}, nil
}

// Clear removes the credential file if present.
func Clear(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
