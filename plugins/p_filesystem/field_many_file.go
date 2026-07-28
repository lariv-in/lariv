package p_filesystem

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
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

func (e FieldManyFile) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
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

	var items bytes.Buffer
	for _, n := range nodes {
		if n.ID != 0 {
			if err := buildFileInfo(n, "", cat, ctx, &items); err != nil {
				return err
			}
		}
	}

	return executeTemplate(w, "field_many_file", struct {
		Classes string
		Items   template.HTML
	}{
		Classes: e.Classes,
		Items:   template.HTML(items.String()),
	})
}
