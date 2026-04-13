package getters

import (
	"context"

	"golang.org/x/exp/constraints"
)

// Context key constants for shared use across packages.
const (
	ContextKeyDB    = "$db"
	ContextKeyError = "$error"
	ContextKeyGet   = "$get"
	ContextKeyIn    = "$in"
)

// Getter defines a common type for fetching data that could be dynamic
type Getter[T any] func(context.Context) (T, error)

type Number interface {
	constraints.Integer | constraints.Float
}
