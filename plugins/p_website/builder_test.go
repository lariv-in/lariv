package p_website

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/views"
	"gorm.io/gorm"
)

func TestIsEditableHTMLName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"page.html", true},
		{"page.HTML", true},
		{"page.htm", true},
		{"page.tmpl", true},
		{"page.css", false},
		{"page", false},
		{"readme.md", false},
	}
	for _, tc := range cases {
		if got := IsEditableHTMLName(tc.name); got != tc.want {
			t.Errorf("IsEditableHTMLName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestBuildPublishedHTML(t *testing.T) {
	out := buildPublishedHTML("<div>Hi</div>", "h1{color:red}")
	if !strings.Contains(out, "<style>") || !strings.Contains(out, "h1{color:red}") {
		t.Fatalf("expected inlined style, got %q", out)
	}
	if !strings.Contains(out, "<div>Hi</div>") {
		t.Fatalf("expected body html, got %q", out)
	}

	full := buildPublishedHTML("<!DOCTYPE html><html><head></head><body>x</body></html>", "a{}")
	if !strings.Contains(full, "</head>") || !strings.Contains(full, "a{}") {
		t.Fatalf("expected style before </head>, got %q", full)
	}
}

func TestCreateBlankPagePatcher(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	ctx := context.WithValue(context.Background(), getters.ContextKeyDB, db)
	req := httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)

	patcher := createBlankPagePatcher{}

	t.Run("requires page when not creating", func(t *testing.T) {
		data, errs := patcher.Patch(views.View{}, req, map[string]any{}, nil)
		if errs["PageID"] == nil {
			t.Fatal("expected PageID error")
		}
		if _, ok := data["CreateNewPage"]; ok {
			t.Fatal("CreateNewPage should be cleared")
		}
	})

	t.Run("rejects bad extension", func(t *testing.T) {
		_, errs := patcher.Patch(views.View{}, req, map[string]any{
			"CreateNewPage": true,
			"NewPageName":   "page.css",
		}, nil)
		if errs["NewPageName"] == nil {
			t.Fatal("expected NewPageName error")
		}
	})

	t.Run("creates blank vnode", func(t *testing.T) {
		data, errs := patcher.Patch(views.View{}, req, map[string]any{
			"CreateNewPage": true,
			"NewPageName":   "home.html",
		}, nil)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		pageID, ok := data["PageID"].(uint)
		if !ok || pageID == 0 {
			t.Fatalf("expected PageID uint, got %#v", data["PageID"])
		}
		var node p_filesystem.VNode
		if err := db.First(&node, pageID).Error; err != nil {
			t.Fatalf("load vnode: %v", err)
		}
		if node.Name != "home.html" {
			t.Fatalf("unexpected name %q", node.Name)
		}
		if node.ParentID != nil {
			t.Fatalf("expected root parent, got %#v", node.ParentID)
		}
		dl, err := node.OpenDownload()
		if err != nil {
			t.Fatalf("OpenDownload: %v", err)
		}
		body, err := io.ReadAll(dl.Reader)
		_ = dl.Reader.Close()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if !strings.Contains(string(body), "<!DOCTYPE html>") {
			t.Fatalf("expected starter html, got %q", string(body))
		}
	})

	t.Run("creates under configured root dir", func(t *testing.T) {
		prev := Config.NewPageRootDir
		Config.NewPageRootDir = "website/pages"
		t.Cleanup(func() { Config.NewPageRootDir = prev })

		data, errs := patcher.Patch(views.View{}, req, map[string]any{
			"CreateNewPage": true,
			"NewPageName":   "about.html",
		}, nil)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		pageID, ok := data["PageID"].(uint)
		if !ok || pageID == 0 {
			t.Fatalf("expected PageID uint, got %#v", data["PageID"])
		}
		var node p_filesystem.VNode
		if err := db.Preload("Parent").First(&node, pageID).Error; err != nil {
			t.Fatalf("load vnode: %v", err)
		}
		if node.Name != "about.html" {
			t.Fatalf("unexpected name %q", node.Name)
		}
		if node.ParentID == nil || node.Parent == nil || node.Parent.Name != "pages" {
			t.Fatalf("expected parent pages dir, got parent=%#v", node.Parent)
		}
		var website p_filesystem.VNode
		if err := db.First(&website, *node.Parent.ParentID).Error; err != nil {
			t.Fatalf("load website parent: %v", err)
		}
		if website.Name != "website" || !website.IsDirectory {
			t.Fatalf("expected website directory ancestor, got %#v", website)
		}
	})
}

func seedRouteWithPage(t *testing.T, db *gorm.DB, name, body string) DBRoute {
	t.Helper()
	node, err := p_filesystem.CreateVNodeFromReader(db, name, strings.NewReader(body), nil)
	if err != nil {
		t.Fatalf("CreateVNodeFromReader: %v", err)
	}
	route := DBRoute{
		Path:     "/seed-" + name,
		PageID:   node.ID,
		Page:     *node,
		IsActive: true,
	}
	if err := db.Create(&route).Error; err != nil {
		t.Fatalf("create route: %v", err)
	}
	route.Page = *node
	return route
}

func TestBuilderLoadAndStore(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	route := seedRouteWithPage(t, db, "edit-me.html", "<html><body>seed</body></html>")

	ctx := context.WithValue(context.Background(), getters.ContextKeyDB, db)
	ctx = context.WithValue(ctx, "dbroute", route)

	t.Run("load seeds from html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/website/1/builder/project/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		loadBuilderProjectHandler(nil).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
		}
		var resp builderLoadResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Data != nil {
			t.Fatalf("expected nil data for first load, got %#v", resp.Data)
		}
		if !strings.Contains(resp.HTML, "seed") {
			t.Fatalf("expected seed html, got %q", resp.HTML)
		}
	})

	t.Run("store project and html", func(t *testing.T) {
		payload := builderProjectPayload{
			Data: map[string]any{
				"pages": []any{
					map[string]any{"name": "Page", "component": "<h1>Built</h1>"},
				},
			},
			HTML: "<h1>Built</h1>",
			CSS:  "h1{color:blue}",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/website/1/builder/project/", bytes.NewReader(body)).WithContext(ctx)
		rec := httptest.NewRecorder()
		storeBuilderProjectHandler(nil).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
		}

		var updated DBRoute
		if err := db.Preload("Page").First(&updated, route.ID).Error; err != nil {
			t.Fatalf("reload route: %v", err)
		}
		if strings.TrimSpace(updated.GrapesProject) == "" {
			t.Fatal("expected GrapesProject to be saved")
		}
		dl, err := updated.Page.OpenDownload()
		if err != nil {
			t.Fatalf("OpenDownload: %v", err)
		}
		content, err := io.ReadAll(dl.Reader)
		_ = dl.Reader.Close()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		published := string(content)
		if !strings.Contains(published, "<h1>Built</h1>") {
			t.Fatalf("expected published html, got %q", published)
		}
		if !strings.Contains(published, "h1{color:blue}") {
			t.Fatalf("expected inlined css, got %q", published)
		}
	})

	t.Run("load returns stored project", func(t *testing.T) {
		var updated DBRoute
		if err := db.Preload("Page").First(&updated, route.ID).Error; err != nil {
			t.Fatalf("reload: %v", err)
		}
		ctx2 := context.WithValue(context.Background(), getters.ContextKeyDB, db)
		ctx2 = context.WithValue(ctx2, "dbroute", updated)
		req := httptest.NewRequest(http.MethodGet, "/website/1/builder/project/", nil).WithContext(ctx2)
		rec := httptest.NewRecorder()
		loadBuilderProjectHandler(nil).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
		}
		var resp builderLoadResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Data == nil {
			t.Fatal("expected stored project data")
		}
	})
}

func TestResolvedAssetsDir(t *testing.T) {
	prevAssets, prevRoot := Config.AssetsDir, Config.NewPageRootDir
	t.Cleanup(func() {
		Config.AssetsDir = prevAssets
		Config.NewPageRootDir = prevRoot
	})

	Config.AssetsDir = "custom/assets"
	Config.NewPageRootDir = "website/"
	if got := Config.ResolvedAssetsDir(); got != "custom/assets" {
		t.Fatalf("got %q", got)
	}

	Config.AssetsDir = ""
	Config.NewPageRootDir = "website/"
	if got := Config.ResolvedAssetsDir(); got != "website/assets" {
		t.Fatalf("got %q", got)
	}

	Config.NewPageRootDir = ""
	if got := Config.ResolvedAssetsDir(); got != "assets" {
		t.Fatalf("got %q", got)
	}
}

func TestBuilderAssetUploadAndPublicServe(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	prevAssets, prevRoot := Config.AssetsDir, Config.NewPageRootDir
	Config.AssetsDir = "website/assets"
	Config.NewPageRootDir = "website/"
	t.Cleanup(func() {
		Config.AssetsDir = prevAssets
		Config.NewPageRootDir = prevRoot
	})

	ctx := context.WithValue(context.Background(), getters.ContextKeyDB, db)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", "hero.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := io.WriteString(part, "fake-png-bytes"); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/website/builder/assets/", &body).WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	builderAssetUploadHandler(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 url, got %#v", resp.Data)
	}
	wantPrefix := "/media/"
	if !strings.HasPrefix(resp.Data[0], wantPrefix) || !strings.HasSuffix(resp.Data[0], "/") {
		t.Fatalf("unexpected url %q", resp.Data[0])
	}

	var nodes []p_filesystem.VNode
	if err := db.Where("is_directory = ?", false).Find(&nodes).Error; err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 file vnode, got %d", len(nodes))
	}
	if !strings.HasPrefix(nodes[0].Name, "hero_") || !strings.HasSuffix(nodes[0].Name, ".png") {
		t.Fatalf("unexpected stored name %q", nodes[0].Name)
	}
	if publicAssetURL(nodes[0].ID) != resp.Data[0] {
		t.Fatalf("url mismatch: got %q want %q", resp.Data[0], publicAssetURL(nodes[0].ID))
	}

	getReq := httptest.NewRequest(http.MethodGet, resp.Data[0], nil).WithContext(ctx)
	getReq.SetPathValue("id", fmt.Sprintf("%d", nodes[0].ID))
	getRec := httptest.NewRecorder()
	publicAssetHandler(nil).ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("public status %d: %s", getRec.Code, getRec.Body.String())
	}
	if getRec.Body.String() != "fake-png-bytes" {
		t.Fatalf("unexpected body %q", getRec.Body.String())
	}
	if !strings.Contains(getRec.Header().Get("Content-Disposition"), "inline") {
		t.Fatalf("expected inline disposition, got %q", getRec.Header().Get("Content-Disposition"))
	}
}
