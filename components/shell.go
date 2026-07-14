package components

import (
	"context"

	"maragu.dev/gomponents"
)

// Shell represents the global base page scaffolding interface (e.g. HTML body wrappers).
// It extends [PageInterface] to define the parent HTML document layout enclosing page-level body structures.
type Shell interface {
	PageInterface
	// Body compiles the core page content wrapper inside the parent HTML document shell structure.
	Body(context.Context) gomponents.Node
}
