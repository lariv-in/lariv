package getters

import (
	"context"
	"testing"
)

func TestSelectMultiRowClass(t *testing.T) {
	got, err := SelectMultiRowClass(Key[uint]("$row.ID"))(context.WithValue(context.Background(), "$row", map[string]any{
		"ID": uint(7),
	}))
	if err != nil {
		t.Fatalf("SelectMultiRowClass returned error: %v", err)
	}
	want := `items.some(item => item.Key === "7") ? 'bg-success text-success-content hover:bg-success border-success' : 'hover:bg-base-200'`
	if got != want {
		t.Fatalf("SelectMultiRowClass() = %q, want %q", got, want)
	}
}
