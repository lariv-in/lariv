package p_seer_websites

import (
	"context"
	"net/url"
	"testing"
)

func TestNormalizeWebsiteURL(t *testing.T) {
	got, err := normalizeWebsiteURL("https://Example.COM/foo/bar?q=1#frag")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/foo/bar?q=1"
	if got.String() != want {
		t.Fatalf("got %q want %q", got.String(), want)
	}
}

func TestURLFailsSSRF(t *testing.T) {
	ctx := context.Background()
	u, _ := url.Parse("http://127.0.0.1/")
	if !urlFailsSSRF(ctx, u) {
		t.Fatal("expected loopback blocked")
	}
	u2, _ := url.Parse("http://8.8.8.8/")
	if urlFailsSSRF(ctx, u2) {
		t.Fatal("expected public resolver allowed")
	}
}
