package getters

import (
	"context"
	"testing"
)

func TestSelectMultiRowClass(t *testing.T) {
	ctx := context.WithValue(context.Background(), "$row", map[string]any{
		"ID": uint(7),
	})
	got, err := SelectMultiRowClass(Static("Courses"), Key[uint]("$row.ID"))(ctx)
	if err != nil {
		t.Fatalf("SelectMultiRowClass returned error: %v", err)
	}
	want := `((Alpine.store('m2mSelections') && Alpine.store('m2mSelections')["Courses"]) || []).some(item => item.Key === "7") ? 'bg-success text-success-content hover:bg-success border-success' : 'hover:bg-base-200'`
	if got != want {
		t.Fatalf("SelectMultiRowClass() = %q, want %q", got, want)
	}
}
