package p_filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// createComponentVNode creates a VNode for a component-uploaded file, using a
// timestamped name to avoid unique-constraint collisions with other parentless nodes.
func createComponentVNode(db *gorm.DB, basePath string, file *multipart.FileHeader) (*VNode, error) {
	ext := filepath.Ext(file.Filename)
	base := strings.TrimSuffix(file.Filename, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", base, time.Now().UnixMilli(), ext)
	parent, err := EnsureDirectoryPath(db, basePath)
	if err != nil {
		slog.Error("failed to ensure directory path for component upload", "error", err, "basePath", basePath)
		return nil, err
	}
	return CreateVNode(db, uniqueName, false, file, parent)
}

func checkFileType(file *multipart.FileHeader, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ct := file.Header.Get("Content-Type")
	for _, ft := range allowed {
		ft = strings.TrimSpace(ft)
		if strings.EqualFold(ft, ext) || strings.EqualFold(ft, ct) {
			return nil
		}
	}
	return fmt.Errorf("file type %q is not allowed", ext)
}

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
			Span(Class("opacity-50"), Text(fmt.Sprintf("(%s)", currentFile.GetFileSize()))),
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

	db := ctx.Value("$db").(*gorm.DB)

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
						Size: n.GetFileSize(),
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
			Class(fmt.Sprintf("file-input file-input-bordered w-full mt-1 %s", e.Classes)),
			Multiple(),
			If(accept != "", Accept(accept)),
		),
	)
}

func (e InputMultiVNode) ParseMultipart(uploadedFiles []*multipart.FileHeader, ctx context.Context) (any, error) {
	db := ctx.Value("$db").(*gorm.DB)

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

func buildFileInfo(v VNode, classes string, ctx context.Context) Node {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = components.DefaultTimeZone
	}
	return Div(
		Class(fmt.Sprintf("flex items-center gap-2 text-sm %s", classes)),
		components.Render(components.Icon{Name: "document"}, ctx),
		Span(Text(v.Name)),
		Span(Class("opacity-50"), Text(fmt.Sprintf("(%s)", v.GetFileSize()))),
		Span(Class("opacity-50"), Text(v.CreatedAt.In(timezone).Format("02 Jan 2006 15:04"))),
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
