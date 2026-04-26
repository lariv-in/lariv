package p_filesystem

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
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
	return new(node), nil
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
		return lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(*node.ParentID)),
		})(ctx)
	}
	return lago.RoutePath("filesystem.ListRoute", nil)(ctx)
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

		redirectURL, err := lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
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

func downloadHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, err := vnodeFromContext(r)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
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

func init() {
	listRoot := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.root", Value: rootVNodeQueryPatcher{}},
			},
		}
	}
	listBrowse := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.browse", Value: browseVNodeQueryPatcher{}},
			},
		}
	}
	listSelectRoot := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.select.root", Value: rootDirectoryQueryPatcher{}},
			},
		}
	}
	listSelectChild := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.select.child", Value: browseDirectoryQueryPatcher{}},
			},
		}
	}
	listMultiRoot := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.multi.root", Value: rootVNodeQueryPatcher{}},
			},
		}
	}
	listMultiChild := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.multi.child", Value: browseVNodeQueryPatcher{}},
			},
		}
	}
	listMoveRoot := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.move-select.root", Value: rootDirectoryQueryPatcher{}},
			},
		}
	}
	listMoveChild := func() views.Layer {
		return views.LayerList[VNode]{
			Key: getters.Static("vnodes"),
			QueryPatchers: views.QueryPatchers[VNode]{
				{Key: "filesystem.move-select.child", Value: browseDirectoryQueryPatcher{}},
			},
		}
	}

	lago.RegistryView.Register("filesystem.ListView",
		lago.GetPageView("filesystem.VNodeTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.list", listRoot()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.BrowseView",
		lago.GetPageView("filesystem.VNodeTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.list", listBrowse()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.DetailView",
		lago.GetPageView("filesystem.VNodeDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}))

	lago.RegistryView.Register("filesystem.CreateView",
		lago.GetPageView("filesystem.VNodeCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.create", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: createHandler,
			}))

	lago.RegistryView.Register("filesystem.CreateChildView",
		lago.GetPageView("filesystem.VNodeCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.create", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: createHandler,
			}))

	lago.RegistryView.Register("filesystem.UpdateView",
		lago.GetPageView("filesystem.VNodeUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
			WithLayer("filesystem.update", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: updateHandler,
			}))

	lago.RegistryView.Register("filesystem.DeleteView",
		lago.GetPageView("filesystem.VNodeDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
			WithLayer("filesystem.delete", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: deleteHandler,
			}))

	lago.RegistryView.Register("filesystem.MoveView",
		lago.GetPageView("filesystem.VNodeMoveForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
			WithLayer("filesystem.move", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: moveHandler,
			}))

	lago.RegistryView.Register("filesystem.MultiUploadView",
		lago.GetPageView("filesystem.VNodeMultiUploadForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.multi_upload", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: multiUploadHandler,
			}))

	lago.RegistryView.Register("filesystem.MultiUploadChildView",
		lago.GetPageView("filesystem.VNodeMultiUploadForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.multi_upload", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: multiUploadHandler,
			}))

	lago.RegistryView.Register("filesystem.SelectView",
		lago.GetPageView("filesystem.ParentSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.list", listSelectRoot()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.SelectChildView",
		lago.GetPageView("filesystem.ParentSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.list", listSelectChild()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.MultiSelectView",
		lago.GetPageView("filesystem.MultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.list", listMultiRoot()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.MultiSelectChildView",
		lago.GetPageView("filesystem.MultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.list", listMultiChild()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.MoveSelectView",
		lago.GetPageView("filesystem.DestinationSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.list", listMoveRoot()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.MoveSelectChildView",
		lago.GetPageView("filesystem.DestinationSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.parent", loadVNodeByPathParamLayer{Param: "parent_id"}).
			WithLayer("filesystem.list", listMoveChild()).
			WithLayer("filesystem.list-enrich", vNodeListEnrichLayer{}))

	lago.RegistryView.Register("filesystem.DownloadView",
		lago.GetPageView("filesystem.VNodeDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("filesystem.node", loadVNodeByPathParamLayer{Param: "id"}).
			WithLayer("filesystem.download", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: downloadHandler,
			}))
}
