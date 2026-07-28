package components

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lariv-in/lariv/getters"
)

func TestFieldMarkdown_Sanitize(t *testing.T) {
	ctx := context.Background()
	unsafeMarkdown := "# Header\n\n<script>alert('xss')</script>\n\nSafe text"

	t.Run("Sanitize default (nil) should sanitize by default", func(t *testing.T) {
		field := FieldMarkdown{
			Getter: getters.Static(unsafeMarkdown),
		}
		var buf bytes.Buffer
		if err := field.Build(EmptyCatalog{}, ctx, &buf); err != nil {
			t.Fatalf("unexpected render error: %v", err)
		}
		rendered := buf.String()
		if strings.Contains(rendered, "<script>") {
			t.Errorf("expected default sanitized output to strip script tag, got: %s", rendered)
		}
		if !strings.Contains(rendered, "Safe text") {
			t.Errorf("expected default sanitized output to retain safe text, got: %s", rendered)
		}
	})

	t.Run("Sanitize enabled (true)", func(t *testing.T) {
		field := FieldMarkdown{
			Getter:   getters.Static(unsafeMarkdown),
			Sanitize: getters.Static(true),
		}
		var buf bytes.Buffer
		if err := field.Build(EmptyCatalog{}, ctx, &buf); err != nil {
			t.Fatalf("unexpected render error: %v", err)
		}
		rendered := buf.String()
		if strings.Contains(rendered, "<script>") {
			t.Errorf("expected sanitized output to strip script tag, got: %s", rendered)
		}
		if !strings.Contains(rendered, "Safe text") {
			t.Errorf("expected sanitized output to retain safe text, got: %s", rendered)
		}
	})

	t.Run("Sanitize explicit false", func(t *testing.T) {
		field := FieldMarkdown{
			Getter:   getters.Static(unsafeMarkdown),
			Sanitize: getters.Static(false),
		}
		var buf bytes.Buffer
		if err := field.Build(EmptyCatalog{}, ctx, &buf); err != nil {
			t.Fatalf("unexpected render error: %v", err)
		}
		rendered := buf.String()
		if !strings.Contains(rendered, "<script>alert('xss')</script>") {
			t.Errorf("expected unsanitized output to contain script tag, got: %s", rendered)
		}
	})
}
