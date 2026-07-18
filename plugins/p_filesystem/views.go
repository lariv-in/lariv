package p_filesystem

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
	"github.com/lariv-in/lariv/views"
	"gorm.io/gorm"
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "sqlstate 23505") ||
		strings.Contains(msg, "violates unique constraint")
}

func filesystemDB(r *http.Request) (*gorm.DB, error) {
	return getters.DBFromContext(r.Context())
}

func parseUintPathValue(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return uint(id), nil
}

type loadVNodeByPathParamLayer struct {
	Param string
}

func (m loadVNodeByPathParamLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := filesystemDB(r)
		if err != nil {
			slog.Error("filesystem: missing db while loading vnode", "param", m.Param, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		id, err := parseUintPathValue(r, m.Param)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		node, err := GetVNodeByID(db, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		n := *node
		n.ResolvedPath = n.GetPath(db)
		n.ListChildrenCount = n.GetChildrenCount(db)
		ctx := context.WithValue(r.Context(), "vnode", n)
		if n.ParentID != nil {
			if p, perr := GetVNodeByID(db, *n.ParentID); perr == nil {
				ctx = context.WithValue(ctx, "vnodeParent", *p)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func vnodeFromContext(r *http.Request) (*VNode, error) {
	node, ok := r.Context().Value("vnode").(VNode)
	if !ok {
		return nil, fmt.Errorf("missing vnode in context")
	}
	return &node, nil
}

func optionalNodeFromValue(db *gorm.DB, value any, fallback *VNode) (*VNode, error) {
	switch v := value.(type) {
	case nil:
		return fallback, nil
	case uint:
		return GetVNodeByID(db, v)
	case string:
		if strings.TrimSpace(v) == "" {
			return fallback, nil
		}
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		return GetVNodeByID(db, uint(id))
	default:
		return nil, fmt.Errorf("unsupported node selector type %T", value)
	}
}

func parentRedirect(ctx context.Context, node *VNode) (string, error) {
	if node != nil && node.ParentID != nil {
		return lariv.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(*node.ParentID)),
		})(ctx)
	}
	return lariv.RoutePath("filesystem.ListRoute", nil)(ctx)
}

type rootVNodeQueryPatcher struct{}

func (rootVNodeQueryPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[VNode]) gorm.ChainInterface[VNode] {
	return q.Order("is_directory DESC").Order("name ASC").Where("parent_id IS NULL")
}

type browseVNodeQueryPatcher struct{}

func (browseVNodeQueryPatcher) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[VNode]) gorm.ChainInterface[VNode] {
	q = q.Order("is_directory DESC").Order("name ASC")
	id, err := parseUintPathValue(r, "parent_id")
	if err != nil {
		return q.Where("1 = 0")
	}
	return q.Where("parent_id = ?", id)
}

type rootDirectoryQueryPatcher struct{}

func (rootDirectoryQueryPatcher) Patch(v views.View, r *http.Request, q gorm.ChainInterface[VNode]) gorm.ChainInterface[VNode] {
	return rootVNodeQueryPatcher{}.Patch(v, r, q).Where("is_directory = ?", true)
}

type browseDirectoryQueryPatcher struct{}

func (browseDirectoryQueryPatcher) Patch(v views.View, r *http.Request, q gorm.ChainInterface[VNode]) gorm.ChainInterface[VNode] {
	return browseVNodeQueryPatcher{}.Patch(v, r, q).Where("is_directory = ?", true)
}

func createHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		name, _ := values["Name"].(string)
		isDirectory, _ := values["IsDirectory"].(bool)
		file, _ := values["File"].(*multipart.FileHeader)

		if isDirectory && strings.TrimSpace(name) == "" {
			name = "New Folder"
		}
		if !isDirectory && file != nil {
			// For file uploads, always derive the persisted node name from the uploaded file.
			name = file.Filename
		}

		var fallbackParent *VNode
		if node, ok := r.Context().Value("vnode").(VNode); ok {
			fallbackParent = &node
		}
		parent, err := optionalNodeFromValue(db, values["ParentID"], fallbackParent)
		if err != nil {
			fieldErrors["ParentID"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		node, err := CreateVNode(db, name, isDirectory, file, parent)
		if err != nil {
			if isUniqueViolation(err) {
				fieldErrors["Name"] = fmt.Errorf("an item with this name already exists here")
			} else {
				fieldErrors["_form"] = err
			}
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		redirectURL, err := parentRedirect(r.Context(), node)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}

func updateHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, err := vnodeFromContext(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		name, _ := values["Name"].(string)
		file, _ := values["File"].(*multipart.FileHeader)
		if strings.TrimSpace(name) == "" {
			fieldErrors["Name"] = fmt.Errorf("name is required")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if err := node.Update(db, name, file); err != nil {
			if isUniqueViolation(err) {
				fieldErrors["Name"] = fmt.Errorf("an item with this name already exists here")
			} else {
				fieldErrors["_form"] = err
			}
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		redirectURL, err := lariv.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(node.ID)),
		})(r.Context())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}

func moveHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, err := vnodeFromContext(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		destination, err := optionalNodeFromValue(db, values["DestinationID"], nil)
		if err != nil {
			fieldErrors["DestinationID"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if err := node.MoveToNode(db, destination); err != nil {
			fieldErrors["DestinationID"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		redirectURL, err := parentRedirect(r.Context(), node)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}

func deleteHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, err := vnodeFromContext(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		redirectURL, err := parentRedirect(r.Context(), node)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if err := node.DeleteTree(db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}

func multiUploadHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		files, _ := values["Files"].([]*multipart.FileHeader)
		if len(files) == 0 {
			fieldErrors["Files"] = fmt.Errorf("at least one file is required")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		var fallbackParent *VNode
		if node, ok := r.Context().Value("vnode").(VNode); ok {
			fallbackParent = &node
		}
		parent, err := optionalNodeFromValue(db, values["ParentID"], fallbackParent)
		if err != nil {
			fieldErrors["ParentID"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		// Preflight duplicate checks so we fail with a precise message and avoid
		// partial uploads when one of many files conflicts.
		seenInBatch := map[string]struct{}{}
		conflicts := []string{}
		for _, file := range files {
			name := sanitizeNodeName(file.Filename)
			if name == "" {
				continue
			}
			if _, ok := seenInBatch[name]; ok {
				conflicts = append(conflicts, name+" (duplicate in selection)")
				continue
			}
			seenInBatch[name] = struct{}{}

			query := db.Model(&VNode{}).Where("name = ? AND is_directory = ?", name, false)
			if parent == nil {
				query = query.Where("parent_id IS NULL")
			} else {
				query = query.Where("parent_id = ?", parent.ID)
			}
			var existingCount int64
			if err := query.Count(&existingCount).Error; err != nil {
				fieldErrors["_form"] = err
				ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
				v.RenderPage(w, r.WithContext(ctx))
				return
			}
			if existingCount > 0 {
				conflicts = append(conflicts, name)
			}
		}
		if len(conflicts) > 0 {
			fieldErrors["Files"] = fmt.Errorf("these files already exist in this location: %s", strings.Join(conflicts, ", "))
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		for _, file := range files {
			if _, err := CreateVNode(db, file.Filename, false, file, parent); err != nil {
				if isUniqueViolation(err) {
					fieldErrors["Files"] = fmt.Errorf("one or more files already exist in this location")
				} else {
					fieldErrors["_form"] = err
				}
				ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
				v.RenderPage(w, r.WithContext(ctx))
				return
			}
		}

		redirectURL, err := parentRedirect(r.Context(), parent)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}

// vNodeListEnrichLayer fills [VNode.ListChildrenCount] for each row (directories) after [views.LayerList].
type vNodeListEnrichLayer struct{}

func (vNodeListEnrichLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ol, ok := r.Context().Value("vnodes").(components.ObjectList[VNode])
		if !ok || len(ol.Items) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, err := filesystemDB(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		var dirIDs []uint
		for _, n := range ol.Items {
			if n.IsDirectory {
				dirIDs = append(dirIDs, n.ID)
			}
		}
		countByID := make(map[uint]int64, len(dirIDs))
		if len(dirIDs) > 0 {
			var aggs []struct {
				ParentID uint  `gorm:"column:parent_id"`
				Cnt      int64 `gorm:"column:cnt"`
			}
			if err := db.Model(&VNode{}).Select("parent_id, count(*) as cnt").Where("parent_id IN (?)", dirIDs).Group("parent_id").Find(&aggs).Error; err != nil {
				slog.Error("filesystem: list enrich: child count query", "error", err)
			} else {
				for _, a := range aggs {
					countByID[a.ParentID] = a.Cnt
				}
			}
		}
		for i := range ol.Items {
			if !ol.Items[i].IsDirectory {
				ol.Items[i].ListChildrenCount = "-"
				continue
			}
			c := countByID[ol.Items[i].ID]
			ol.Items[i].ListChildrenCount = fmt.Sprintf("%d items", c)
		}
		ctx := context.WithValue(r.Context(), "vnodes", ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addVNodeToZip(db *gorm.DB, archive *zip.Writer, node VNode, relativePath string) error {
	if node.IsDirectory {
		// Ensure directory path ends with a slash in ZIP
		dirPath := relativePath
		if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
			dirPath += "/"
		}
		if dirPath != "" {
			header := &zip.FileHeader{
				Name:     dirPath,
				Method:   zip.Store,
				Modified: node.UpdatedAt,
			}
			_, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
		}

		var children []VNode
		var query *gorm.DB
		if node.ID != 0 {
			query = db.Where("parent_id = ?", node.ID)
		} else {
			query = db.Where("parent_id IS NULL")
		}
		if err := query.Find(&children).Error; err != nil {
			return err
		}
		for i := range children {
			childPath := filepath.Join(relativePath, children[i].Name)
			// archive/zip paths must always use forward slashes, even on Windows
			childPath = strings.ReplaceAll(childPath, "\\", "/")
			if err := addVNodeToZip(db, archive, children[i], childPath); err != nil {
				return err
			}
		}
		return nil
	}

	download, err := node.OpenDownload()
	if err != nil {
		return err
	}
	defer download.Reader.Close()

	header := &zip.FileHeader{
		Name:     relativePath,
		Method:   zip.Deflate,
		Modified: node.UpdatedAt,
	}
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(writer, download.Reader); err != nil {
		return err
	}
	return nil
}

func downloadRootHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="root.zip"`)

		archive := zip.NewWriter(w)
		defer archive.Close()

		rootNode := VNode{
			Name:        "root",
			IsDirectory: true,
		}

		if err := addVNodeToZip(db, archive, rootNode, ""); err != nil {
			slog.Error("filesystem: failed zipping root directory", "error", err)
		}
	})
}

func downloadHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, err := vnodeFromContext(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if node.IsDirectory {
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", node.Name+".zip"))

			archive := zip.NewWriter(w)
			defer archive.Close()

			if err := addVNodeToZip(db, archive, *node, ""); err != nil {
				slog.Error("filesystem: failed zipping directory", "id", node.ID, "error", err)
			}
			return
		}

		download, err := node.OpenDownload()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer download.Reader.Close()

		w.Header().Set("Content-Type", download.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", download.Size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", download.Filename))
		if _, err := io.Copy(w, download.Reader); err != nil {
			slog.Error("filesystem: failed writing download response", "id", node.ID, "error", err)
		}
	})
}

func filesystemLayerListRoot() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.root", Value: rootVNodeQueryPatcher{}},
		},
	}
}

func filesystemLayerListBrowse() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.browse", Value: browseVNodeQueryPatcher{}},
		},
	}
}

func filesystemLayerListSelectRoot() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.select.root", Value: rootDirectoryQueryPatcher{}},
		},
	}
}

func filesystemLayerListSelectChild() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.select.child", Value: browseDirectoryQueryPatcher{}},
		},
	}
}

func filesystemLayerListMultiRoot() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.multi.root", Value: rootVNodeQueryPatcher{}},
		},
	}
}

func filesystemLayerListMultiChild() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.multi.child", Value: browseVNodeQueryPatcher{}},
		},
	}
}

func filesystemLayerListMoveRoot() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.move-select.root", Value: rootDirectoryQueryPatcher{}},
		},
	}
}

func filesystemLayerListMoveChild() views.Layer {
	return views.LayerList[VNode]{
		Key: getters.Static("vnodes"),
		QueryPatchers: views.QueryPatchers[VNode]{
			{Key: "filesystem.move-select.child", Value: browseDirectoryQueryPatcher{}},
		},
	}
}

func pluginViews() lariv.PluginFeatures[*views.View] {
	return lariv.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			{Key: "filesystem.ListView", Value: lariv.GetPageView("filesystem.VNodeTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.list", filesystemLayerListRoot()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.BrowseView", Value: lariv.GetPageView("filesystem.VNodeTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.list", filesystemLayerListBrowse()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.DetailView", Value: lariv.GetPageView("filesystem.VNodeDetail").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"})},
			{Key: "filesystem.CreateView", Value: lariv.GetPageView("filesystem.VNodeCreateForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.create", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: createHandler,
				})},
			{Key: "filesystem.CreateChildView", Value: lariv.GetPageView("filesystem.VNodeCreateForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.create", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: createHandler,
				})},
			{Key: "filesystem.UpdateView", Value: lariv.GetPageView("filesystem.VNodeUpdateForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
				WithLayer("filesystem.update", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: updateHandler,
				})},
			{Key: "filesystem.DeleteView", Value: lariv.GetPageView("filesystem.VNodeDeleteForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
				WithLayer("filesystem.delete", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: deleteHandler,
				})},
			{Key: "filesystem.MoveView", Value: lariv.GetPageView("filesystem.VNodeMoveForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
				WithLayer("filesystem.move", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: moveHandler,
				})},
			{Key: "filesystem.MultiUploadView", Value: lariv.GetPageView("filesystem.VNodeMultiUploadForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.multi_upload", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: multiUploadHandler,
				})},
			{Key: "filesystem.MultiUploadChildView", Value: lariv.GetPageView("filesystem.VNodeMultiUploadForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.multi_upload", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: multiUploadHandler,
				})},
			{Key: "filesystem.ZipUploadView", Value: lariv.GetPageView("filesystem.VNodeZipUploadForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.zip_upload", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: zipUploadHandler,
				})},
			{Key: "filesystem.ZipUploadChildView", Value: lariv.GetPageView("filesystem.VNodeZipUploadForm").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.zip_upload", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: zipUploadHandler,
				})},
			{Key: "filesystem.SelectView", Value: lariv.GetPageView("filesystem.ParentSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.list", filesystemLayerListSelectRoot()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.SelectChildView", Value: lariv.GetPageView("filesystem.ParentSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.list", filesystemLayerListSelectChild()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.FileSelectView", Value: lariv.GetPageView("filesystem.FileSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.list", filesystemLayerListMultiRoot()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.FileSelectChildView", Value: lariv.GetPageView("filesystem.FileSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.list", filesystemLayerListMultiChild()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.MultiSelectView", Value: lariv.GetPageView("filesystem.MultiSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.list", filesystemLayerListMultiRoot()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.MultiSelectChildView", Value: lariv.GetPageView("filesystem.MultiSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.list", filesystemLayerListMultiChild()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.MoveSelectView", Value: lariv.GetPageView("filesystem.DestinationSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.list", filesystemLayerListMoveRoot()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.MoveSelectChildView", Value: lariv.GetPageView("filesystem.DestinationSelectionTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
				WithLayer("filesystem.list", filesystemLayerListMoveChild()).
				WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{})},
			{Key: "filesystem.DownloadView", Value: lariv.GetPageView("filesystem.VNodeDetail").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
				WithLayer("filesystem.download", views.MethodLayer{
					Method:  http.MethodGet,
					Handler: downloadHandler,
				})},
			{Key: "filesystem.DownloadRootView", Value: lariv.GetPageView("filesystem.VNodeTable").
				WithLayer("p_users.auth", p_users.AuthenticationLayer{}).
				WithLayer("filesystem.download_root", views.MethodLayer{
					Method:  http.MethodGet,
					Handler: downloadRootHandler,
				})},
		},
	}
}

// chatUploadHandler accepts multipart files, creates VNodes, and returns a JSON
// array of {id, name} objects for use by the chat interface file upload button.
func chatUploadHandler(w http.ResponseWriter, r *http.Request) {
	db, err := filesystemDB(r)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, `{"error":"invalid multipart form"}`, http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["Files"]
	if len(files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}
	type nodeResult struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	results := make([]nodeResult, 0, len(files))
	for _, fh := range files {
		node, err := createComponentVNode(db, "", fh)
		if err != nil {
			slog.Error("chatUploadHandler: failed to create vnode", "file", fh.Filename, "error", err)
			continue
		}
		results = append(results, nodeResult{ID: node.ID, Name: node.Name})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func zipUploadHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := filesystemDB(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		zipFileHeader, _ := values["ZipFile"].(*multipart.FileHeader)
		if zipFileHeader == nil {
			fieldErrors["ZipFile"] = fmt.Errorf("zip file is required")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		// Verify it is actually a zip file
		if !strings.HasSuffix(strings.ToLower(zipFileHeader.Filename), ".zip") {
			fieldErrors["ZipFile"] = fmt.Errorf("uploaded file is not a zip archive")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		var fallbackParent *VNode
		if node, ok := r.Context().Value("vnode").(VNode); ok {
			fallbackParent = &node
		}
		parent, err := optionalNodeFromValue(db, values["ParentID"], fallbackParent)
		if err != nil {
			fieldErrors["ParentID"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		if parent != nil && !parent.IsDirectory {
			fieldErrors["ParentID"] = fmt.Errorf("target must be a directory")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		src, err := zipFileHeader.Open()
		if err != nil {
			fieldErrors["ZipFile"] = fmt.Errorf("failed to open uploaded zip file: %w", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		defer src.Close()

		zipReader, err := zip.NewReader(src, zipFileHeader.Size)
		if err != nil {
			fieldErrors["ZipFile"] = fmt.Errorf("failed to parse zip file: %w", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		var createdFiles []string
		txErr := db.Transaction(func(tx *gorm.DB) error {
			// Delete direct children recursively
			var children []VNode
			var query *gorm.DB
			if parent != nil {
				query = tx.Where("parent_id = ?", parent.ID)
			} else {
				query = tx.Where("parent_id IS NULL")
			}
			if err := query.Find(&children).Error; err != nil {
				return fmt.Errorf("failed to list old contents: %w", err)
			}
			for i := range children {
				if err := children[i].DeleteTree(tx); err != nil {
					return fmt.Errorf("failed to delete old item %q: %w", children[i].Name, err)
				}
			}

			// Extract ZIP entries
			for _, f := range zipReader.File {
				cleanPath := filepath.Clean(f.Name)
				if strings.HasPrefix(cleanPath, "../") || strings.Contains(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
					continue
				}
				cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
				parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
				if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
					continue
				}

				currentParent := parent
				var dirParts []string
				if f.FileInfo().IsDir() {
					dirParts = parts
				} else {
					dirParts = parts[:len(parts)-1]
				}

				for _, dirName := range dirParts {
					dirName = sanitizeNodeName(dirName)
					if dirName == "" {
						continue
					}
					var existing VNode
					var q *gorm.DB
					if currentParent != nil {
						q = tx.Where("parent_id = ? AND name = ?", currentParent.ID, dirName)
					} else {
						q = tx.Where("parent_id IS NULL AND name = ?", dirName)
					}
					err := q.First(&existing).Error
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							newDir := &VNode{
								Name:        dirName,
								IsDirectory: true,
							}
							if currentParent != nil {
								newDir.ParentID = &currentParent.ID
							}
							if err := gorm.G[VNode](tx).Create(context.Background(), newDir); err != nil {
								return fmt.Errorf("failed to create directory %q: %w", dirName, err)
							}
							currentParent = newDir
						} else {
							return fmt.Errorf("failed to check existing directory %q: %w", dirName, err)
						}
					} else {
						if !existing.IsDirectory {
							return fmt.Errorf("path conflict: %q exists but is not a directory", dirName)
						}
						currentParent = &existing
					}
				}

				if !f.FileInfo().IsDir() {
					fileName := sanitizeNodeName(parts[len(parts)-1])
					if fileName == "" {
						continue
					}
					rc, err := f.Open()
					if err != nil {
						return fmt.Errorf("failed to open file in zip: %w", err)
					}
					ext := filepath.Ext(fileName)
					storedPath, err := Store.SaveFromReader(rc, ext)
					rc.Close()
					if err != nil {
						return fmt.Errorf("failed to save stored file %q: %w", fileName, err)
					}
					createdFiles = append(createdFiles, storedPath)

					node := &VNode{
						Name:        fileName,
						IsDirectory: false,
						FilePath:    storedPath,
					}
					if currentParent != nil {
						node.ParentID = &currentParent.ID
					}
					if err := gorm.G[VNode](tx).Create(context.Background(), node); err != nil {
						return fmt.Errorf("failed to create database file node %q: %w", fileName, err)
					}
				}
			}
			return nil
		})

		if txErr != nil {
			slog.Error("zipUploadHandler: transaction failed, cleaning up stored files", "error", txErr)
			for _, fPath := range createdFiles {
				if delErr := Store.Delete(fPath); delErr != nil {
					slog.Error("zipUploadHandler: failed to clean up physical file", "path", fPath, "error", delErr)
				}
			}
			fieldErrors["_form"] = txErr
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		var redirectURL string
		if parent != nil {
			redirectURL, err = lariv.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
				"parent_id": getters.Any(getters.Static(parent.ID)),
			})(r.Context())
		} else {
			redirectURL, err = lariv.RoutePath("filesystem.ListRoute", nil)(r.Context())
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
	})
}
