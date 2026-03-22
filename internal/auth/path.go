package auth

import (
	"os"
	"path/filepath"
)

// DefaultTokenDir is ~/.weixin-bot
func DefaultTokenDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".weixin-bot"), nil
}

// DefaultTokenPath returns ~/.weixin-bot/credentials.json
func DefaultTokenPath() (string, error) {
	dir, err := DefaultTokenDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// ResolveTokenPath returns the given path or the default credentials path.
func ResolveTokenPath(tokenPath string) (string, error) {
	if tokenPath != "" {
		return filepath.Clean(tokenPath), nil
	}
	return DefaultTokenPath()
}
