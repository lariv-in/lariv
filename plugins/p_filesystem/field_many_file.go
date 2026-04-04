package p_filesystem

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldManyFile struct {
	components.Page
	VNode   getters.Getter[[]VNode]
	Classes string
}

func (e FieldManyFile) GetKey() string {
	return e.Key
}

func (e FieldManyFile) GetRoles() []string {
	return e.Roles
}

func (e FieldManyFile) Build(ctx context.Context) Node {
	if e.VNode == nil {
		return nil
	}

	nodes, err := e.VNode(ctx)
	if err != nil {
		slog.Error("FieldManyFile getter failed", "error", err, "key", e.Key)
		return nil
	}
	if len(nodes) == 0 {
		return nil
	}

	var items []Node
	for _, n := range nodes {
		if n.ID != 0 {
			items = append(items, buildFileInfo(n, "", ctx))
		}
	}

	return Div(Class(e.Classes), Group(items))
}
