package p_website

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func TestPluginGrapesJSBlocksLoadFromEmbed(t *testing.T) {
	features := pluginGrapesJSBlocks()
	if len(features.Entries) < 30 {
		t.Fatalf("expected at least 30 blocks, got %d", len(features.Entries))
	}

	byKey := map[string]lariv.GrapesJSBlock{}
	for _, e := range features.Entries {
		byKey[e.Key] = e.Value
	}
	for _, key := range []string{
		"p_website.section",
		"p_website.2-columns",
		"p_website.3-columns",
		"p_website.card",
		"p_website.accordion",
		"p_website.dotlottie",
		"p_website.number-counter",
	} {
		block, ok := byKey[key]
		if !ok {
			t.Fatalf("missing block %q", key)
		}
		var content string
		if err := json.Unmarshal(block.Content, &content); err != nil {
			t.Fatalf("%s content: %v", key, err)
		}
		if content == "" {
			t.Fatalf("%s content empty", key)
		}
	}

	section := byKey["p_website.section"]
	var cat string
	if err := json.Unmarshal(section.Category, &cat); err != nil || cat != "Layout" {
		t.Fatalf("section category: %v %q", err, cat)
	}
}

func TestGrapesJSBlocksJSONPayload(t *testing.T) {
	entries := pluginGrapesJSBlocks().Entries
	raw, err := grapesJSBlocksJSON(&entries)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var payload []map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload) != len(entries) {
		t.Fatalf("expected %d payload entries, got %d", len(entries), len(payload))
	}
	if payload[0]["id"] != "p_website.section" {
		t.Fatalf("first id: %#v", payload[0]["id"])
	}
	if _, ok := payload[0]["content"].(string); !ok {
		t.Fatalf("content should be string in JSON, got %#v", payload[0]["content"])
	}

	empty, err := grapesJSBlocksJSON((*[]registry.Pair[string, lariv.GrapesJSBlock])(nil))
	if err != nil {
		t.Fatalf("nil encode: %v", err)
	}
	if string(empty) != "[]" {
		t.Fatalf("nil encode want [], got %s", empty)
	}
}

func TestPluginGrapesJSComponents(t *testing.T) {
	features := pluginGrapesJSComponents()
	if len(features.Entries) != 26 {
		t.Fatalf("expected 26 components, got %d", len(features.Entries))
	}
	raw, err := grapesJSComponentsJSON(&features.Entries)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var payload []map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload) != 26 {
		t.Fatalf("payload len %d", len(payload))
	}
	foundDot := false
	for _, item := range payload {
		if item["id"] == "p_website.dotlottie" {
			foundDot = true
			if item["view"] == nil {
				t.Fatal("dotlottie should include view.onRender")
			}
		}
	}
	if !foundDot {
		t.Fatal("missing p_website.dotlottie")
	}

	empty, err := grapesJSComponentsJSON((*[]registry.Pair[string, lariv.GrapesJSComponent])(nil))
	if err != nil {
		t.Fatalf("nil encode: %v", err)
	}
	if string(empty) != "[]" {
		t.Fatalf("nil encode want [], got %s", empty)
	}
}

func TestPluginGrapesJSTraits(t *testing.T) {
	features := pluginGrapesJSTraits()
	if len(features.Entries) != 1 {
		t.Fatalf("expected 1 trait, got %d", len(features.Entries))
	}
	if features.Entries[0].Key != "p_website.src-url" {
		t.Fatalf("trait key: %s", features.Entries[0].Key)
	}
	raw, err := grapesJSTraitsJSON(&features.Entries)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var payload []map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload[0]["id"] != "p_website.src-url" {
		t.Fatalf("id: %#v", payload[0]["id"])
	}
	if _, ok := payload[0]["createInput"].(string); !ok {
		t.Fatalf("createInput should be string, got %#v", payload[0]["createInput"])
	}
}

func TestInjectDotLottieScript(t *testing.T) {
	plain := `<!DOCTYPE html><html><body><h1>Hi</h1></body></html>`
	if got := injectDotLottieScript(plain); got != plain {
		t.Fatalf("unused page should be unchanged")
	}

	withAnim := `<!DOCTYPE html><html><body><dotlottie-wc src="x.lottie"></dotlottie-wc></body></html>`
	injected := injectDotLottieScript(withAnim)
	if !strings.Contains(injected, dotLottieScriptAttr) {
		t.Fatalf("expected script attr in %s", injected)
	}
	if !strings.Contains(injected, dotLottieCDNURL) {
		t.Fatalf("expected CDN url in %s", injected)
	}
	if strings.Count(injected, "dotlottie-wc.js") != 1 {
		t.Fatalf("expected exactly one script src, got %s", injected)
	}
	if idxScript, idxBody := strings.Index(injected, "dotlottie-wc.js"), strings.Index(strings.ToLower(injected), "</body>"); idxScript < 0 || idxBody < 0 || idxScript > idxBody {
		t.Fatalf("script should appear before </body>")
	}

	again := injectDotLottieScript(injected)
	if again != injected {
		t.Fatalf("inject should be idempotent")
	}
}
