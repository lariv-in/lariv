package lariv

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGrapesJSTraitHelpers(t *testing.T) {
	create := GrapesJSTraitCreateInputJS("return document.createElement('input');")
	var body string
	if err := json.Unmarshal(create, &body); err != nil || !strings.Contains(body, "createElement") {
		t.Fatalf("createInput: %v %q", err, body)
	}

	onEvent := GrapesJSTraitOnEventJS("component.addAttributes({ src: elInput.value });")
	if err := json.Unmarshal(onEvent, &body); err != nil || !strings.Contains(body, "addAttributes") {
		t.Fatalf("onEvent: %v %q", err, body)
	}

	onUpdate := GrapesJSTraitOnUpdateJS("elInput.value = component.getAttributes().src || '';")
	if err := json.Unmarshal(onUpdate, &body); err != nil || !strings.Contains(body, "getAttributes") {
		t.Fatalf("onUpdate: %v %q", err, body)
	}

	tpl := GrapesJSTraitTemplateInputString(`<div data-input></div>`)
	var html string
	if err := json.Unmarshal(tpl, &html); err != nil || html == "" {
		t.Fatalf("templateInput: %v %q", err, html)
	}
}

func TestGrapesJSTraitJSONRoundTrip(t *testing.T) {
	trait := GrapesJSTrait{
		NoLabel:      true,
		EventCapture: []string{"input", "change"},
		CreateInput:  GrapesJSTraitCreateInputJS("return document.createElement('input');"),
		OnEvent:      GrapesJSTraitOnEventJS("component.addAttributes({ src: elInput.value });"),
		OnUpdate:     GrapesJSTraitOnUpdateJS("elInput.value = component.getAttributes().src || '';"),
	}
	raw, err := json.Marshal(trait)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded["noLabel"] != true {
		t.Fatalf("noLabel: %#v", decoded["noLabel"])
	}
	if _, ok := decoded["createInput"].(string); !ok {
		t.Fatalf("createInput should be string, got %#v", decoded["createInput"])
	}
}
