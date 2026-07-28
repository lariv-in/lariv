package p_website

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/views"
	"gorm.io/gorm"
)

// publicAssetURL returns the public path that streams a VNode by id.
func publicAssetURL(id uint) string {
	return fmt.Sprintf("/media/%d/", id)
}

func createBuilderAssetVNode(db *gorm.DB, file *multipart.FileHeader) (*p_filesystem.VNode, error) {
	ext := filepath.Ext(file.Filename)
	base := strings.TrimSuffix(file.Filename, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", base, time.Now().UnixMilli(), ext)
	parent, err := p_filesystem.EnsureDirectoryPath(db, Config.ResolvedAssetsDir())
	if err != nil {
		return nil, err
	}
	return p_filesystem.CreateVNode(db, uniqueName, false, file, parent)
}

func multipartFiles(r *http.Request) []*multipart.FileHeader {
	if r.MultipartForm == nil {
		return nil
	}
	for _, key := range []string{"files[]", "files", "Files"} {
		if files := r.MultipartForm.File[key]; len(files) > 0 {
			return files
		}
	}
	// GrapesJS may append an index suffix; collect any files* keys.
	var out []*multipart.FileHeader
	for key, files := range r.MultipartForm.File {
		if strings.HasPrefix(strings.ToLower(key), "files") {
			out = append(out, files...)
		}
	}
	return out
}

// builderAssetUploadHandler accepts GrapesJS AssetManager multipart uploads,
// stores each file as a p_filesystem VNode, and returns {"data":[url,...]}.
func builderAssetUploadHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			http.Error(w, `{"error":"invalid multipart form"}`, http.StatusBadRequest)
			return
		}

		files := multipartFiles(r)
		urls := make([]string, 0, len(files))
		for _, fh := range files {
			node, err := createBuilderAssetVNode(db, fh)
			if err != nil {
				slog.Error("builder asset upload: create vnode", "file", fh.Filename, "error", err)
				http.Error(w, `{"error":"failed to store asset"}`, http.StatusInternalServerError)
				return
			}
			urls = append(urls, publicAssetURL(node.ID))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"data": urls}); err != nil {
			slog.Error("builder asset upload: encode response", "error", err)
		}
	})
}

// publicAssetHandler streams a VNode for use as a public <img src> / CSS url.
func publicAssetHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			http.NotFound(w, r)
			return
		}

		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		node, err := p_filesystem.GetVNodeByID(db, uint(id))
		if err != nil || node.IsDirectory {
			http.NotFound(w, r)
			return
		}

		download, err := node.OpenDownload()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer download.Reader.Close()

		w.Header().Set("Content-Type", download.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", download.Size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", download.Filename))
		if _, err := io.Copy(w, download.Reader); err != nil {
			slog.Error("public asset: write response", "id", node.ID, "error", err)
		}
	})
}
