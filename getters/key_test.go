package getters

import (
	"context"
	"errors"
	"testing"
)

func TestKeyTraversesMapStringError(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyError, map[string]error{
		"Password": errors.New("invalid email or password"),
	})
	g := Key[error]("$error.Password")
	v, err := g(ctx)
	if err != nil {
		t.Fatalf("getter error: %v", err)
	}
	if v == nil || v.Error() != "invalid email or password" {
		t.Fatalf("got %#v want invalid email or password", v)
	}
}
