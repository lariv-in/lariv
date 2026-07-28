package p_filesystem

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
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

func (e InputVNode) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
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

	data := struct {
		Classes  string
		Label    string
		Name     string
		Required bool
		Accept   string
		HasFile  bool
		Icon     template.HTML
		FileName string
		FileSize string
	}{
		Classes:  e.Classes,
		Label:    e.Label,
		Name:     e.Name,
		Required: e.Required && currentFile == nil,
		Accept:   accept,
	}

	if currentFile != nil {
		icon, err := components.RenderHTML(components.Icon{Name: "document"}, cat, ctx)
		if err != nil {
			return err
		}
		data.HasFile = true
		data.Icon = icon
		data.FileName = currentFile.Name
		data.FileSize = currentFile.FileSizeDisplay()
	}

	return executeTemplate(w, "input_vnode", data)
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
