package p_filesystem

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
)

func buildFileInfo(v VNode, classes string, cat components.Catalog, ctx context.Context, w io.Writer) error {
	rowClass := fmt.Sprintf("flex items-center gap-2 text-sm %s", classes)
	icon, err := components.RenderHTML(components.Icon{Name: "document"}, cat, ctx)
	if err != nil {
		return err
	}

	detailURL, err := lariv.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(v.ID)),
	})(ctx)
	if err != nil {
		slog.Error("buildFileInfo detail route resolution failed", "error", err, "vnodeID", v.ID)
		return executeTemplate(w, "field_file", struct {
			Href     string
			RowClass string
			Icon     template.HTML
			Name     string
			Size     string
		}{
			RowClass: rowClass,
			Icon:     icon,
			Name:     v.Name,
			Size:     v.FileSizeDisplay(),
		})
	}

	return executeTemplate(w, "field_file", struct {
		Href     string
		RowClass string
		Icon     template.HTML
		Name     string
		Size     string
	}{
		Href:     detailURL,
		RowClass: rowClass,
		Icon:     icon,
		Name:     v.Name,
		Size:     v.FileSizeDisplay(),
	})
}

type FieldFile struct {
	components.Page
	VNode   getters.Getter[VNode]
	Classes string
}

func (e FieldFile) GetKey() string {
	return e.Key
}

func (e FieldFile) GetRoles() []string {
	return e.Roles
}

func (e FieldFile) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	if e.VNode == nil {
		return nil
	}

	v, err := e.VNode(ctx)
	if err != nil {
		slog.Error("FieldFile getter failed", "error", err, "key", e.Key)
		return nil
	}
	if v.ID == 0 {
		return nil
	}

	return buildFileInfo(v, e.Classes, cat, ctx, w)
}
