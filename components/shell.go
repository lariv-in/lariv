package components

import (
	"context"
	"io"
)

// Shell represents the global base page scaffolding interface (e.g. HTML body wrappers).
// It extends [PageInterface] to define the parent HTML document layout enclosing page-level body structures.
type Shell interface {
	PageInterface
	// Body compiles the core page content wrapper inside the parent HTML document shell structure.
	Body(cat Catalog, ctx context.Context, w io.Writer) error
}
