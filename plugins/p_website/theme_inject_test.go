package p_website

import (
	"strings"
	"testing"

	"github.com/lariv-in/lariv"
)

func TestPluginGrapesJSThemes(t *testing.T) {
	features := pluginGrapesJSThemes()
	if len(features.Entries) != 6 {
		t.Fatalf("expected 6 themes, got %d", len(features.Entries))
	}
	byKey := map[string]lariv.GrapesJSTheme{}
	for _, e := range features.Entries {
		byKey[e.Key] = e.Value
	}
	def, ok := byKey[defaultThemeID]
	if !ok {
		t.Fatal("missing default theme")
	}
	if !strings.Contains(def.CSS, "--lariv-ink") {
		t.Fatal("expected embedded theme CSS variables")
	}
	if len(def.Stylesheets) == 0 {
		t.Fatal("expected font stylesheet")
	}
	for _, key := range []string{
		"p_website.mvp",
		"p_website.tacit",
		"p_website.pico",
		"p_website.water",
		"p_website.marx",
	} {
		theme, ok := byKey[key]
		if !ok {
			t.Fatalf("missing theme %q", key)
		}
		if len(theme.Stylesheets) == 0 {
			t.Fatalf("%s: expected CDN stylesheet", key)
		}
	}
}

func TestInjectThemeAssets(t *testing.T) {
	theme := lariv.GrapesJSTheme{
		Label:       "Default",
		CSS:         ".gjs-hero{color:red}",
		Stylesheets: []string{"https://example.com/fonts.css"},
	}

	plain := `<!DOCTYPE html><html><head></head><body><h1>Hi</h1></body></html>`
	if got := injectThemeAssets(plain, "", theme); got != plain {
		t.Fatalf("empty theme should leave doc unchanged")
	}

	injected := injectThemeAssets(plain, defaultThemeID, theme)
	if !strings.Contains(injected, `data-lariv-theme="`+defaultThemeID+`"`) {
		t.Fatalf("missing theme marker: %s", injected)
	}
	if !strings.Contains(injected, "https://example.com/fonts.css") {
		t.Fatalf("missing stylesheet: %s", injected)
	}
	if !strings.Contains(injected, ".gjs-hero{color:red}") {
		t.Fatalf("missing css: %s", injected)
	}

	replaced := injectThemeAssets(injected, "other.theme", lariv.GrapesJSTheme{CSS: ".x{}"})
	if strings.Contains(replaced, defaultThemeID) {
		t.Fatalf("old theme should be stripped: %s", replaced)
	}
	if !strings.Contains(replaced, `data-lariv-theme="other.theme"`) || !strings.Contains(replaced, ".x{}") {
		t.Fatalf("new theme missing: %s", replaced)
	}
}

func TestBuildPublishedHTMLWithTheme(t *testing.T) {
	out := buildPublishedHTMLWithTheme("<div>Hi</div>", "h1{}", defaultThemeID, lariv.GrapesJSTheme{
		CSS: ".gjs-button{background:teal}",
	})
	if !strings.Contains(out, "h1{}") || !strings.Contains(out, ".gjs-button{background:teal}") {
		t.Fatalf("expected editor + theme css, got %q", out)
	}
}

func TestGrapesJSThemesJSONPayload(t *testing.T) {
	entries := pluginGrapesJSThemes().Entries
	raw, err := grapesJSThemesJSON(&entries)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(string(raw), `"id":"p_website.default"`) {
		t.Fatalf("payload: %s", raw)
	}
}
