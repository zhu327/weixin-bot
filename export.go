package weixinbot

import "github.com/zhu327/weixin-bot/internal/auth"

func credentialsFromData(d *auth.Data) *Credentials {
	if d == nil {
		return nil
	}
	return &Credentials{
		Token:     d.Token,
		BaseURL:   d.BaseURL,
		AccountID: d.AccountID,
		UserID:    d.UserID,
	}
}

// LoadCredentials reads stored credentials from disk (or nil if missing).
func LoadCredentials(tokenPath string) (*Credentials, error) {
	p, err := auth.ResolveTokenPath(tokenPath)
	if err != nil {
		return nil, err
	}
	d, err := auth.Load(p)
	if err != nil {
		return nil, err
	}
	return credentialsFromData(d), nil
}

// ClearCredentials removes the credential file if present.
func ClearCredentials(tokenPath string) error {
	p, err := auth.ResolveTokenPath(tokenPath)
	if err != nil {
		return err
	}
	return auth.Clear(p)
}
