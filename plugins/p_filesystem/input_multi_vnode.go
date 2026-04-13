package p_filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputMultiVNode struct {
	components.Page
	VNode            getters.Getter[[]VNode]
	Label            string
	Name             string
	Required         bool
	Classes          string
	AllowedFiletypes []string
	Path             getters.Getter[string]
}

type multiVNodeItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
}

func (e InputMultiVNode) GetKey() string {
	return e.Key
}

func (e InputMultiVNode) GetRoles() []string {
	return e.Roles
}

func (e InputMultiVNode) Build(ctx context.Context) Node {
	var items []multiVNodeItem
	if e.VNode != nil {
		nodes, err := e.VNode(ctx)
		if err != nil {
			slog.Error("InputMultiVNode getter failed", "error", err, "key", e.Key)
		} else {
			for _, n := range nodes {
				if n.ID != 0 {
					items = append(items, multiVNodeItem{
						ID:   strconv.FormatUint(uint64(n.ID), 10),
						Name: n.Name,
						Size: n.FileSizeDisplay(),
					})
				}
			}
		}
	}
	if items == nil {
		items = []multiVNodeItem{}
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		slog.Error("InputMultiVNode items marshal failed", "error", err, "key", e.Key)
		itemsJSON = []byte("[]")
	}

	accept := strings.Join(e.AllowedFiletypes, ",")

	alpineData := fmt.Sprintf(`{
		files: %s,
		removeFile(id) {
			this.files = this.files.filter(f => f.id !== id)
		}
	}`, itemsJSON)

	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Attr("x-data", alpineData),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Template(
			Attr("x-for", "file in files"),
			Attr(":key", "file.id"),
			Div(
				Class("flex items-center gap-2 text-sm my-1 rounded-lg bg-base-200 px-2 py-1"),
				Input(Type("hidden"), Name(e.Name), Attr(":value", "file.id")),
				components.Render(components.Icon{Name: "document"}, ctx),
				Span(Class("flex-1 min-w-0 truncate"), Attr("x-text", "file.name")),
				Span(Class("opacity-50 shrink-0"), Attr("x-text", "'(' + file.size + ')'")),
				Button(
					Type("button"),
					Class("btn btn-ghost btn-square btn-xs shrink-0"),
					Attr("@click.stop", "removeFile(file.id)"),
					Attr("aria-label", "Remove"),
					components.Render(components.Icon{Name: "x-mark"}, ctx),
				),
			),
		),
		Input(
			Type("file"),
			Name(e.Name),
			Class(fmt.Sprintf("file-input file-input-bordered w-full %s", e.Classes)),
			Multiple(),
			If(accept != "", Accept(accept)),
		),
	)
}

func (e InputMultiVNode) ParseMultipart(uploadedFiles []*multipart.FileHeader, ctx context.Context) (any, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Collect kept IDs from the hidden inputs (available when $request is in context).
	keptIDs := map[uint]struct{}{}
	if r, ok := ctx.Value("$request").(*http.Request); ok && r.MultipartForm != nil {
		for _, raw := range r.MultipartForm.Value[e.Name] {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			id, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				continue
			}
			keptIDs[uint(id)] = struct{}{}
		}
	}

	// Delete previously linked VNodes that were removed.
	if e.VNode != nil {
		previous, err := e.VNode(ctx)
		if err == nil {
			for i := range previous {
				if previous[i].ID == 0 {
					continue
				}
				if _, kept := keptIDs[previous[i].ID]; !kept {
					if err := previous[i].DeleteTree(db); err != nil {
						slog.Error("InputMultiVNode failed to delete removed file", "error", err, "id", previous[i].ID)
					}
				}
			}
		}
	}

	// Validate and create new VNodes for uploaded files.
	for _, file := range uploadedFiles {
		if err := checkFileType(file, e.AllowedFiletypes); err != nil {
			return components.AssociationIDs{Field: e.Name}, err
		}

		path := ""
		if e.Path != nil {
			var err error
			path, err = e.Path(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get path: %w", err)
			}
		}

		node, err := createComponentVNode(db, path, file)
		if err != nil {
			return components.AssociationIDs{Field: e.Name}, err
		}
		keptIDs[node.ID] = struct{}{}
	}

	allIDs := make([]uint, 0, len(keptIDs))
	for id := range keptIDs {
		allIDs = append(allIDs, id)
	}

	if e.Required && len(allIDs) == 0 {
		return components.AssociationIDs{Field: e.Name, IDs: allIDs}, fmt.Errorf("at least one file is required")
	}

	return components.AssociationIDs{Field: e.Name, IDs: allIDs}, nil
}

func (e InputMultiVNode) Parse(v any, _ context.Context) (any, error) {
	return nil, nil
}

func (e InputMultiVNode) GetName() string {
	return e.Name
}
