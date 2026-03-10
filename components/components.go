package components

import (
	"context"

	"maragu.dev/gomponents"
)

type PageInterface interface {
	Build(context.Context) gomponents.Node
}

// Shell is implemented by top-level page shells (scaffolds) that wrap content in <html>/<head>/<body>.
// Body returns only the inner content of <body>, which is used for HTMX boosted requests
// to avoid re-sending the full <html> document.
type Shell interface {
	PageInterface
	Body(context.Context) gomponents.Node
}
