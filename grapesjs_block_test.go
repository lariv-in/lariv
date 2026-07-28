package lariv

import (
	"encoding/json"
	"testing"
)

func TestGrapesJSBlockContentHelpers(t *testing.T) {
	content := GrapesJSBlockContentString("<div>hi</div>")
	var s string
	if err := json.Unmarshal(content, &s); err != nil {
		t.Fatalf("unmarshal content string: %v", err)
	}
	if s != "<div>hi</div>" {
		t.Fatalf("unexpected content %q", s)
	}

	obj, err := GrapesJSBlockContentJSON(map[string]any{"type": "image"})
	if err != nil {
		t.Fatalf("ContentJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(obj, &m); err != nil {
		t.Fatalf("unmarshal content object: %v", err)
	}
	if m["type"] != "image" {
		t.Fatalf("unexpected object %#v", m)
	}

	cat := GrapesJSBlockCategoryString("Layout")
	var catStr string
	if err := json.Unmarshal(cat, &catStr); err != nil || catStr != "Layout" {
		t.Fatalf("category: %v %q", err, catStr)
	}

	onBool := GrapesJSBlockOnClickBool(true)
	var b bool
	if err := json.Unmarshal(onBool, &b); err != nil || !b {
		t.Fatalf("onClick bool: %v %v", err, b)
	}

	onJS := GrapesJSBlockOnClickJS("return true;")
	var body string
	if err := json.Unmarshal(onJS, &body); err != nil || body != "return true;" {
		t.Fatalf("onClick js: %v %q", err, body)
	}
}

func TestGrapesJSBlockJSONRoundTrip(t *testing.T) {
	block := GrapesJSBlock{
		Label:    "Section",
		Category: GrapesJSBlockCategoryString("Layout"),
		Content:  GrapesJSBlockContentString("<section></section>"),
		Activate: true,
	}
	raw, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}
	if decoded["label"] != "Section" {
		t.Fatalf("label: %#v", decoded["label"])
	}
	if decoded["category"] != "Layout" {
		t.Fatalf("category should decode as string, got %#v", decoded["category"])
	}
	if decoded["content"] != "<section></section>" {
		t.Fatalf("content should decode as string, got %#v", decoded["content"])
	}
}
