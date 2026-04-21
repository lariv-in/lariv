package p_seer_opensky

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOpenSkyCredentialsFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "opensky.json")
	content := `{"client_id":"abc","client_secret":"secret"}`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	id, sec, err := loadOpenSkyCredentialsFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc" || sec != "secret" {
		t.Fatalf("got %q %q", id, sec)
	}
}

func TestLoadOpenSkyCredentialsFileCamelCase(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "opensky.json")
	content := `{"clientId":"x","clientSecret":"y"}`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	id, sec, err := loadOpenSkyCredentialsFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if id != "x" || sec != "y" {
		t.Fatalf("got %q %q", id, sec)
	}
}

func TestLoadOpenSkyCredentialsFileMissingKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "opensky.json")
	if err := os.WriteFile(p, []byte(`{"client_id":"only"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := loadOpenSkyCredentialsFile(p)
	if err == nil {
		t.Fatal("want error")
	}
}
