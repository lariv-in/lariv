package p_filesystem

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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

func (e FieldPhoto) Build(ctx context.Context) Node {
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

	downloadURL, err := lago.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
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

	return Img(
		Src(downloadURL),
		Alt(alt),
		Class(e.Classes),
	)
}
