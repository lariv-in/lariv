package p_filesystem

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func buildFileInfo(v VNode, classes string, ctx context.Context) Node {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = components.DefaultTimeZone
	}
	rowClass := fmt.Sprintf("flex items-center gap-2 text-sm %s", classes)
	children := []Node{
		components.Render(components.Icon{Name: "document"}, ctx),
		Span(Text(v.Name)),
		Span(Class("opacity-50"), Text(fmt.Sprintf("(%s)", v.FileSizeDisplay()))),
	}

	detailURL, err := lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(v.ID)),
	})(ctx)
	if err != nil {
		slog.Error("buildFileInfo detail route resolution failed", "error", err, "vnodeID", v.ID)
		return Div(Class(rowClass), Group(children))
	}

	return A(
		Href(detailURL),
		Class(rowClass+" link link-hover"),
		Group(children),
	)
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

func (e FieldFile) Build(ctx context.Context) Node {
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

	return buildFileInfo(v, e.Classes, ctx)
}
