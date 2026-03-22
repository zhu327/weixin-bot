package auth

import (
	"errors"
	"os"
	"path/filepath"
)

// DefaultTokenDir picks the first writable directory among: ~/.weixin-bot, ./.weixin-bot (cwd),
// then $TMP/weixin-bot. Cwd is preferred over the system temp dir so credentials survive typical
// /tmp cleanup when home is unavailable (e.g. some containers).
func DefaultTokenDir() (string, error) {
	var candidates []string
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates, filepath.Join(home, ".weixin-bot"))
	}
	if wd, err := os.Getwd(); err == nil && wd != "" {
		candidates = append(candidates, filepath.Join(wd, ".weixin-bot"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "weixin-bot"))
	for _, dir := range candidates {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			continue
		}
		fi, err := os.Stat(dir)
		if err != nil || !fi.IsDir() {
			continue
		}
		f, err := os.CreateTemp(dir, ".write-test-*")
		if err != nil {
			continue
		}
		_ = f.Close()
		_ = os.Remove(f.Name())
		return dir, nil
	}
	return "", errors.New("weixinbot/auth: no writable directory for token storage")
}

// DefaultTokenPath returns credentials.json under [DefaultTokenDir].
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
