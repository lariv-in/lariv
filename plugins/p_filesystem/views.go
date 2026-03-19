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

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
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
	db, ok := r.Context().Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("missing database in context")
	}
	return db, nil
}

func parseUintPathValue(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return uint(id), nil
}

func loadVNodeMiddleware(param string) views.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			db, err := filesystemDB(r)
			if err != nil {
				slog.Error("filesystem: missing db while loading vnode", "param", param, "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			id, err := parseUintPathValue(r, param)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			node, err := GetVNodeByID(db, id)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), "vnode", *node)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
		return lago.GetterRoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.GetterAny(getters.GetterStatic(*node.ParentID)),
		})(ctx)
	}
	return lago.GetterRoutePath("filesystem.ListRoute", nil)(ctx)
}

func rootVNodeQuery(_ *views.View, _ *http.Request, query *gorm.DB) *gorm.DB {
	return ListChildrenForParent(query, nil)
}

func browseVNodeQuery(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	id, err := parseUintPathValue(r, "parent_id")
	if err != nil {
		return query.Where("1 = 0")
	}
	return ListChildrenForParent(query, &id)
}

func rootDirectoryQuery(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	return rootVNodeQuery(v, r, query).Where("is_directory = ?", true)
}

func browseDirectoryQuery(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	return browseVNodeQuery(v, r, query).Where("is_directory = ?", true)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		node, err := CreateVNode(db, name, isDirectory, file, parent)
		if err != nil {
			if isUniqueViolation(err) {
				fieldErrors["Name"] = fmt.Errorf("an item with this name already exists here")
			} else {
				fieldErrors["_form"] = err
			}
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		redirectURL, err := parentRedirect(r.Context(), node)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if err := node.Update(db, name, file); err != nil {
			if isUniqueViolation(err) {
				fieldErrors["Name"] = fmt.Errorf("an item with this name already exists here")
			} else {
				fieldErrors["_form"] = err
			}
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		redirectURL, err := lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterStatic(node.ID)),
		})(r.Context())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if err := node.MoveToNode(db, destination); err != nil {
			fieldErrors["DestinationID"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		redirectURL, err := parentRedirect(r.Context(), node)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		var fallbackParent *VNode
		if node, ok := r.Context().Value("vnode").(VNode); ok {
			fallbackParent = &node
		}
		parent, err := optionalNodeFromValue(db, values["ParentID"], fallbackParent)
		if err != nil {
			fieldErrors["ParentID"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
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
				v.RenderWithErrors(w, r, fieldErrors, values)
				return
			}
			if existingCount > 0 {
				conflicts = append(conflicts, name)
			}
		}
		if len(conflicts) > 0 {
			fieldErrors["Files"] = fmt.Errorf("these files already exist in this location: %s", strings.Join(conflicts, ", "))
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		for _, file := range files {
			if _, err := CreateVNode(db, file.Filename, false, file, parent); err != nil {
				if isUniqueViolation(err) {
					fieldErrors["Files"] = fmt.Errorf("one or more files already exist in this location")
				} else {
					fieldErrors["_form"] = err
				}
				v.RenderWithErrors(w, r, fieldErrors, values)
				return
			}
		}

		redirectURL, err := parentRedirect(r.Context(), parent)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
	lago.RegistryView.Register("filesystem.ListView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.VNodeTable")).
			WithQueryPatcher("filesystem.root", rootVNodeQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.BrowseView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.VNodeTable")).
			WithQueryPatcher("filesystem.browse", browseVNodeQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.DetailView",
		lago.GetPageView("filesystem.VNodeDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.node", loadVNodeMiddleware("id")))

	lago.RegistryView.Register("filesystem.CreateView",
		lago.GetPageView("filesystem.VNodeCreateForm").
			WithMethod(http.MethodPost, createHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.CreateChildView",
		lago.GetPageView("filesystem.VNodeCreateForm").
			WithMethod(http.MethodPost, createHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.UpdateView",
		lago.GetPageView("filesystem.VNodeUpdateForm").
			WithMethod(http.MethodPost, updateHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.node", loadVNodeMiddleware("id")))

	lago.RegistryView.Register("filesystem.DeleteView",
		lago.GetPageView("filesystem.VNodeDeleteForm").
			WithMethod(http.MethodPost, deleteHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.node", loadVNodeMiddleware("id")))

	lago.RegistryView.Register("filesystem.MoveView",
		lago.GetPageView("filesystem.VNodeMoveForm").
			WithMethod(http.MethodPost, moveHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.node", loadVNodeMiddleware("id")))

	lago.RegistryView.Register("filesystem.MultiUploadView",
		lago.GetPageView("filesystem.VNodeMultiUploadForm").
			WithMethod(http.MethodPost, multiUploadHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.MultiUploadChildView",
		lago.GetPageView("filesystem.VNodeMultiUploadForm").
			WithMethod(http.MethodPost, multiUploadHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.SelectView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.ParentSelectionTable")).
			WithQueryPatcher("filesystem.select.root", rootDirectoryQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.SelectChildView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.ParentSelectionTable")).
			WithQueryPatcher("filesystem.select.child", browseDirectoryQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.MultiSelectView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.MultiSelectionTable")).
			WithQueryPatcher("filesystem.multi.root", rootVNodeQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.MultiSelectChildView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.MultiSelectionTable")).
			WithQueryPatcher("filesystem.multi.child", browseVNodeQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.MoveSelectView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.DestinationSelectionTable")).
			WithQueryPatcher("filesystem.move-select.root", rootDirectoryQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("filesystem.MoveSelectChildView",
		views.ListView[VNode]("vnodes")(lago.GetPageView("filesystem.DestinationSelectionTable")).
			WithQueryPatcher("filesystem.move-select.child", browseDirectoryQuery).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.parent", loadVNodeMiddleware("parent_id")))

	lago.RegistryView.Register("filesystem.DownloadView",
		lago.GetPageView("filesystem.VNodeDetail").
			WithMethod(http.MethodGet, downloadHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("filesystem.node", loadVNodeMiddleware("id")))
}
