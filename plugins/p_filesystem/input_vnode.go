package p_filesystem

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputVNode struct {
	components.Page
	VNode            getters.Getter[VNode]
	Label            string
	Name             string
	Required         bool
	Classes          string
	AllowedFiletypes []string
	Path             getters.Getter[string]
}

func (e InputVNode) GetKey() string {
	return e.Key
}

func (e InputVNode) GetRoles() []string {
	return e.Roles
}

func (e InputVNode) Build(ctx context.Context) Node {
	var currentFile *VNode
	if e.VNode != nil {
		v, err := e.VNode(ctx)
		if err != nil {
			slog.Error("InputVNode getter failed", "error", err, "key", e.Key)
		} else if v.ID != 0 {
			currentFile = &v
		}
	}

	accept := strings.Join(e.AllowedFiletypes, ",")

	var fileInfo Node
	if currentFile != nil {
		fileInfo = Div(Class("flex items-center gap-2 text-sm my-1"),
			components.Render(components.Icon{Name: "document"}, ctx),
			Span(Text(currentFile.Name)),
			Span(Class("opacity-50"), Text(fmt.Sprintf("(%s)", currentFile.FileSizeDisplay()))),
		)
	}

	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		If(fileInfo != nil, fileInfo),
		Input(
			Type("file"),
			Name(e.Name),
			Class(fmt.Sprintf("file-input file-input-bordered w-full %s", e.Classes)),
			If(e.Required && currentFile == nil, Required()),
			If(accept != "", Accept(accept)),
		),
	)
}

func (e InputVNode) ParseMultipart(files []*multipart.FileHeader, ctx context.Context) (any, error) {
	if len(files) == 0 {
		return nil, nil
	}

	file := files[0]

	if err := checkFileType(file, e.AllowedFiletypes); err != nil {
		return nil, err
	}

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Delete the previously linked VNode if one exists.
	if e.VNode != nil {
		old, err := e.VNode(ctx)
		if err == nil && old.ID != 0 {
			if err := old.DeleteTree(db); err != nil {
				slog.Error("InputVNode failed to delete previous file", "error", err, "id", old.ID)
			}
		}
	}

	if e.Path == nil {
		node, err := createComponentVNode(db, "", file)
		if err != nil {
			return nil, err
		}

		return node.ID, nil
	}

	path, err := e.Path(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
	}

	node, err := createComponentVNode(db, path, file)
	if err != nil {
		return nil, err
	}

	return node.ID, nil
}

func (e InputVNode) Parse(v any, ctx context.Context) (any, error) {
	return nil, nil
}

func (e InputVNode) GetName() string {
	return e.Name
}
