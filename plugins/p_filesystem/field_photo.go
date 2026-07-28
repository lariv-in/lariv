package p_filesystem

import (
	"context"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
)

type FieldPhoto struct {
	components.Page
	VNode   getters.Getter[VNode]
	Alt     string
	Classes string
}

func (e FieldPhoto) GetKey() string {
	return e.Key
}

func (e FieldPhoto) GetRoles() []string {
	return e.Roles
}

func (e FieldPhoto) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	if e.VNode == nil {
		return nil
	}

	v, err := e.VNode(ctx)
	if err != nil {
		slog.Error("FieldPhoto getter failed", "error", err, "key", e.Key)
		return nil
	}
	if v.ID == 0 {
		return nil
	}

	downloadURL, err := lariv.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(v.ID)),
	})(ctx)
	if err != nil {
		slog.Error("FieldPhoto route resolution failed", "error", err, "key", e.Key)
		return nil
	}

	alt := e.Alt
	if alt == "" {
		alt = v.Name
	}

	return executeTemplate(w, "field_photo", struct {
		Src     string
		Alt     string
		Classes string
	}{
		Src:     downloadURL,
		Alt:     alt,
		Classes: e.Classes,
	})
}
