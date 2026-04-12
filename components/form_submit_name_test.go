package components

import (
	"context"
	"strings"
	"testing"

	"github.com/lariv-in/lago/getters"
)

func TestFormListenBoostedPostRendersNameGuard(t *testing.T) {
	html := renderNode(t, FormListenBoostedPost{
		Name:      getters.Static("example.submit"),
		ActionURL: getters.Static("/submit"),
		Children:  []PageInterface{},
	}.Build(context.Background()))

	if !strings.Contains(html, `d.name !==`) {
		t.Fatalf("expected name guard in boosted-post listener, got %s", html)
	}
	if !strings.Contains(html, "example.submit") {
		t.Fatalf("expected listener name in rendered html, got %s", html)
	}
	if strings.Index(html, `d.name !==`) > strings.Index(html, `evt.stopPropagation()`) {
		t.Fatalf("expected stopPropagation after name guard, got %s", html)
	}
	if !strings.Contains(html, "targetPath") || !strings.Contains(html, "formPath") {
		t.Fatalf("expected action-path guard in boosted-post listener, got %s", html)
	}
	if !strings.Contains(html, "root.contains(f)") {
		t.Fatalf("expected subtree containment guard in boosted-post listener, got %s", html)
	}
	if !strings.Contains(html, "lagoPostPending") {
		t.Fatalf("expected in-flight submit guard in boosted-post listener, got %s", html)
	}
}

func TestButtonModalFormThreadsNameIntoGetAndGuard(t *testing.T) {
	html := renderNode(t, ButtonModalForm{
		Name:        getters.Static("example.modal"),
		Url:         getters.Static("/modal"),
		FormPostURL: getters.Static("/modal/post"),
		ModalUID:    "example-modal",
		Label:       "Open",
	}.Build(context.Background()))

	if !strings.Contains(html, `hx-get="/modal?name=example.modal"`) {
		t.Fatalf("expected modal opener to carry the name query param, got %s", html)
	}
	if !strings.Contains(html, "example.modal") {
		t.Fatalf("expected modal opener to render the configured name, got %s", html)
	}
	if !strings.Contains(html, `d.name !==`) {
		t.Fatalf("expected modal submit listener to guard on name, got %s", html)
	}
	if !strings.Contains(html, `/modal/post?name=example.modal`) {
		t.Fatalf("expected modal POST URL to carry name query for $get on validation re-render, got %s", html)
	}
	if strings.Index(html, `d.name !==`) > strings.Index(html, `evt.stopPropagation()`) {
		t.Fatalf("expected stopPropagation after the name guard, got %s", html)
	}
	if !strings.Contains(html, "lagoPostPending") {
		t.Fatalf("expected in-flight submit guard in modal form listener, got %s", html)
	}
}

func TestFormBubblingUsesRequestName(t *testing.T) {
	ctx := context.WithValue(context.Background(), "$get", map[string]any{
		"name": "example.modal",
	})

	html := renderNode(t, FormComponent[struct{}]{
		Attr: getters.FormBubbling(getters.Key[string]("$get.name")),
		ChildrenAction: []PageInterface{
			ButtonSubmit{Label: "Save"},
		},
	}.Build(ctx))

	if !strings.Contains(html, "example.modal") {
		t.Fatalf("expected bubbled form to embed the request name, got %s", html)
	}
}
