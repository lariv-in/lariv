// Package getters defines the core Getter type and a suite of utility functions
// for fetching, transforming, and composing dynamic values.
//
// The central type is [Getter], which represents a deferred or dynamic value
// that can be resolved given a [context.Context].
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
