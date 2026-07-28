package lariv

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGrapesJSComponentHelpers(t *testing.T) {
	js := GrapesJSComponentIsComponentJS("return el.tagName === 'DIV';")
	var body string
	if err := json.Unmarshal(js, &body); err != nil || body != "return el.tagName === 'DIV';" {
		t.Fatalf("isComponent js: %v %q", err, body)
	}

	b := GrapesJSComponentIsComponentBool(false)
	var flag bool
	if err := json.Unmarshal(b, &flag); err != nil || flag {
		t.Fatalf("isComponent bool: %v %v", err, flag)
	}

	tag := GrapesJSComponentIsComponentTag("INPUT")
	if err := json.Unmarshal(tag, &body); err != nil {
		t.Fatalf("isComponent tag: %v", err)
	}
	if !strings.Contains(body, "INPUT") {
		t.Fatalf("expected tag in body, got %q", body)
	}

	model, err := GrapesJSComponentModelJSON(map[string]any{
		"defaults": map[string]any{"tagName": "div"},
	})
	if err != nil {
		t.Fatalf("ModelJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(model, &m); err != nil {
		t.Fatalf("unmarshal model: %v", err)
	}
	defaults, ok := m["defaults"].(map[string]any)
	if !ok || defaults["tagName"] != "div" {
		t.Fatalf("unexpected model %#v", m)
	}

	view := GrapesJSComponentMustViewJSON(map[string]any{"tagName": "div"})
	var v map[string]any
	if err := json.Unmarshal(view, &v); err != nil || v["tagName"] != "div" {
		t.Fatalf("view: %v %#v", err, v)
	}
}

func TestGrapesJSComponentJSONRoundTrip(t *testing.T) {
	comp := GrapesJSComponent{
		Extend:      "text",
		IsComponent: GrapesJSComponentIsComponentTag("h1"),
		Model: GrapesJSComponentMustModelJSON(map[string]any{
			"defaults": map[string]any{"tagName": "h1"},
		}),
	}
	raw, err := json.Marshal(comp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["extend"] != "text" {
		t.Fatalf("extend: %#v", decoded["extend"])
	}
	if _, ok := decoded["isComponent"].(string); !ok {
		t.Fatalf("isComponent should be string body, got %#v", decoded["isComponent"])
	}
}
