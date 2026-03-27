package forms

import (
	"context"
	"testing"
)

func TestFormIDFromPathContext(t *testing.T) {
	tests := []struct {
		name string
		path map[string]any
		want uint
		ok   bool
	}{
		{"form_id", map[string]any{"form_id": "7"}, 7, true},
		{"empty form_id", map[string]any{"form_id": ""}, 0, false},
		{"missing", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.path != nil {
				ctx = context.WithValue(ctx, "$path", tt.path)
			}
			got, ok := formIDFromPathContext(ctx)
			if ok != tt.ok || got != tt.want {
				t.Errorf("formIDFromPathContext() = (%d, %v), want (%d, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
