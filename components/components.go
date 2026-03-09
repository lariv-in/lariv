package components

import (
	"context"
	"maragu.dev/gomponents"
)

type PageInterface interface {
	Build(context.Context) gomponents.Node
}
