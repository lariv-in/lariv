package getters

import (
	"context"
	"testing"
	"time"
)

func TestRefRoundTripsWithDeref(t *testing.T) {
	ctx := context.Background()
	orig := 42
	g := Static(&orig)
	v, err := Deref(g)(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Fatalf("Deref: got %d want 42", v)
	}

	rg := Ref(Static(99))
	p, err := rg(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 99 {
		t.Fatalf("Ref: got %v want 99", p)
	}
	if got, err := Deref(rg)(ctx); err != nil || got != 99 {
		t.Fatalf("Deref(Ref): got %v err %v", got, err)
	}
}

func TestRefDurationForInputDuration(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyIn, map[string]any{
		"Duration": 3 * time.Hour,
	})
	g := Ref(Key[time.Duration]("$in.Duration"))
	p, err := g(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 3*time.Hour {
		t.Fatalf("got %v want 3h", p)
	}
}
