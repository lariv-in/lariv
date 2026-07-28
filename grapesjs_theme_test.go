package lariv

import (
	"encoding/json"
	"testing"
)

func TestGrapesJSThemeJSONRoundTrip(t *testing.T) {
	theme := GrapesJSTheme{
		Label:       "Default",
		CSS:         ".gjs-hero { color: red; }",
		Stylesheets: []string{"https://example.com/fonts.css"},
	}
	raw, err := json.Marshal(theme)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GrapesJSTheme
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Label != "Default" || decoded.CSS == "" || len(decoded.Stylesheets) != 1 {
		t.Fatalf("unexpected %#v", decoded)
	}
}
