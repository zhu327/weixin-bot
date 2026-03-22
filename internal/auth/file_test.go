package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "credentials.json")
	d := &Data{Token: "t", BaseURL: "https://example.com", AccountID: "a", UserID: "u"}
	if err := Save(p, d); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != d.Token || got.BaseURL != d.BaseURL {
		t.Fatalf("%+v vs %+v", got, d)
	}
}

func TestLoadSnakeCase(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "credentials.json")
	content := `{
  "token": "tok",
  "base_url": "https://x.com",
  "account_id": "acc",
  "user_id": "usr"
}
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://x.com" || got.UserID != "usr" {
		t.Fatalf("%+v", got)
	}
}

func TestLoadMissing(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}
