package components

import (
	"context"

	"maragu.dev/gomponents"
)

type Shell interface {
	PageInterface
	Body(context.Context) gomponents.Node
}
