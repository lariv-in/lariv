package p_website

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testDBForViews sets up SQLite in-memory DB and runs migrations.
func testDBForViews(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}

	// Migrate both VNode and DBRoute
	if err := db.AutoMigrate(&p_filesystem.VNode{}, &DBRoute{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS filesystem_nodes_parent_name_dir_uidx ON filesystem_nodes (COALESCE(parent_id, 0), name, is_directory)",
	).Error; err != nil {
		t.Fatalf("creating unique index failed: %v", err)
	}

	return db
}

func withTempStorageForViews(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	previous := p_filesystem.Store
	p_filesystem.Store = &p_filesystem.LocalFilestore{BaseDir: dir}
	t.Cleanup(func() {
		p_filesystem.Store = previous
	})
	return dir
}

func uploadHeaderForViews(t *testing.T, fieldName, fileName, body string) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := io.WriteString(part, body); err != nil {
		t.Fatalf("writing multipart body failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(4 * 1024 * 1024); err != nil {
		t.Fatalf("ParseMultipartForm failed: %v", err)
	}
	files := req.MultipartForm.File[fieldName]
	if len(files) == 0 {
		t.Fatalf("no file header created for %s", fieldName)
	}
	return files[0]
}

// buildHandler assembles the full HTTP handler pipeline from registered plugins.
func buildHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	config := lariv.LarivConfig{}
	corePlugin := lariv.CorePlugin(db, config)
	plugins := []registry.Pair[string, lariv.Plugin]{
		corePlugin,
		GetPlugin(),
	}
	lariv.BuildAllRegistries(plugins)
	var handler http.Handler = lariv.GetRouter(config)
	for _, layer := range *lariv.RegistryLayer.AllStable() {
		handler = layer.Value.Next(handler)
	}
	return handler
}

func TestDynamicWebsiteRouting(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	// Create a database virtual node for a template file
	fileHeader := uploadHeaderForViews(t, "File", "index.html", "Hello World Dynamic Route!")
	vnode, err := p_filesystem.CreateVNode(db, "", false, fileHeader, nil)
	if err != nil {
		t.Fatalf("failed to create VNode template file: %v", err)
	}

	// Insert an active DBRoute pointing to this template
	route := DBRoute{
		Path:     "/hello",
		PageID:   vnode.ID,
		IsActive: true,
	}
	if err := db.Create(&route).Error; err != nil {
		t.Fatalf("failed to create DBRoute: %v", err)
	}

	// Also insert an inactive DBRoute
	inactiveRoute := DBRoute{
		Path:     "/inactive",
		PageID:   vnode.ID,
		IsActive: false,
	}
	if err := db.Create(&inactiveRoute).Error; err != nil {
		t.Fatalf("failed to create inactive DBRoute: %v", err)
	}

	handler := buildHandler(t, db)

	// 1. Test valid active route
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", res.StatusCode, string(body))
	}
	if !strings.Contains(string(body), "Hello World Dynamic Route!") {
		t.Errorf("expected body to contain template contents, got: %s", string(body))
	}

	// 2. Test 404 for inactive route
	reqInactive := httptest.NewRequest("GET", "/inactive", nil)
	recInactive := httptest.NewRecorder()
	handler.ServeHTTP(recInactive, reqInactive)

	if recInactive.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected inactive route to yield 404, got %d", recInactive.Result().StatusCode)
	}

	// 3. Test 404 for non-existent route
	reqMissing := httptest.NewRequest("GET", "/missing", nil)
	recMissing := httptest.NewRecorder()
	handler.ServeHTTP(recMissing, reqMissing)

	if recMissing.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected non-existent route to yield 404, got %d", recMissing.Result().StatusCode)
	}
}

func TestRoutesListRendersOK(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	// Seed a VNode and a DBRoute
	fileHeader := uploadHeaderForViews(t, "File", "page.html", "<h1>Hello</h1>")
	vnode, err := p_filesystem.CreateVNode(db, "", false, fileHeader, nil)
	if err != nil {
		t.Fatalf("failed to create VNode: %v", err)
	}
	if err := db.Create(&DBRoute{Path: "/hello", PageID: vnode.ID, IsActive: true}).Error; err != nil {
		t.Fatalf("failed to create DBRoute: %v", err)
	}

	handler := buildHandler(t, db)

	// The list endpoint is under /website/ (AppURL). No auth is set up, so it
	// should redirect to login (3xx) rather than panic or 500.
	req := httptest.NewRequest("GET", AppURL, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	// Without a valid session the auth layer will redirect (302/303) to login.
	if res.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(res.Body)
		t.Errorf("expected redirect or 200, got 500. Body: %s", string(body))
	}
}

func TestRoutesCreateAndListRoundtrip(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	// Seed a VNode that the form will reference
	fileHeader := uploadHeaderForViews(t, "File", "tmpl.html", "template content")
	vnode, err := p_filesystem.CreateVNode(db, "", false, fileHeader, nil)
	if err != nil {
		t.Fatalf("failed to create VNode: %v", err)
	}

	// Directly create a DBRoute via the DB to verify the model round-trips
	newRoute := DBRoute{
		Path:     "/test-create",
		PageID:   vnode.ID,
		IsActive: true,
	}
	if err := db.Create(&newRoute).Error; err != nil {
		t.Fatalf("failed to create DBRoute: %v", err)
	}

	// Verify it was persisted
	var found DBRoute
	if err := db.Where("path = ?", "/test-create").First(&found).Error; err != nil {
		t.Fatalf("route not found after creation: %v", err)
	}
	if found.PageID != vnode.ID {
		t.Errorf("expected PageID %d, got %d", vnode.ID, found.PageID)
	}
	if !found.IsActive {
		t.Error("expected IsActive to be true")
	}
}

func TestRoutesDetailAndUpdateRoundtrip(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	// Seed resources
	fileHeader := uploadHeaderForViews(t, "File", "tmpl.html", "template content")
	vnode, err := p_filesystem.CreateVNode(db, "", false, fileHeader, nil)
	if err != nil {
		t.Fatalf("failed to create VNode: %v", err)
	}
	route := &DBRoute{Path: "/original-path", PageID: vnode.ID, IsActive: true}
	if err := db.Create(route).Error; err != nil {
		t.Fatalf("failed to create DBRoute: %v", err)
	}

	// Update via DB
	route.Path = "/updated-path"
	if err := db.Save(route).Error; err != nil {
		t.Fatalf("failed to update DBRoute: %v", err)
	}

	// Verify updated value persisted
	var updated DBRoute
	if err := db.First(&updated, route.ID).Error; err != nil {
		t.Fatalf("failed to fetch updated route: %v", err)
	}
	if updated.Path != "/updated-path" {
		t.Errorf("expected path /updated-path, got %s", updated.Path)
	}
}

func TestRoutesDeleteRoundtrip(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	fileHeader := uploadHeaderForViews(t, "File", "tmpl.html", "template content")
	vnode, err := p_filesystem.CreateVNode(db, "", false, fileHeader, nil)
	if err != nil {
		t.Fatalf("failed to create VNode: %v", err)
	}
	route := &DBRoute{Path: "/to-delete", PageID: vnode.ID, IsActive: true}
	if err := db.Create(route).Error; err != nil {
		t.Fatalf("failed to create DBRoute: %v", err)
	}

	// Soft-delete via DB
	if err := db.Delete(route).Error; err != nil {
		t.Fatalf("failed to delete DBRoute: %v", err)
	}

	// Verify it no longer appears in active queries
	var count int64
	db.Model(&DBRoute{}).Where("path = ?", "/to-delete").Count(&count)
	if count != 0 {
		t.Errorf("expected route to be deleted, but count=%d", count)
	}
}

func TestRoutesCRUDHTTPStatusCodes(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	handler := buildHandler(t, db)

	endpoints := []struct {
		method string
		path   string
		form   url.Values
	}{
		{"GET", AppURL, nil},
		{"GET", AppURL + "create/", nil},
		{"GET", AppURL + "999/", nil},
		{"GET", AppURL + "999/edit/", nil},
		{"GET", AppURL + "999/delete/", nil},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var req *http.Request
			if ep.form != nil {
				req = httptest.NewRequest(ep.method, ep.path, strings.NewReader(ep.form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(ep.method, ep.path, nil)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			res := rec.Result()

			// Auth layer will redirect unauthenticated requests; we should never get 500.
			if res.StatusCode == http.StatusInternalServerError {
				body, _ := io.ReadAll(res.Body)
				t.Errorf("%s %s returned 500. Body: %s", ep.method, ep.path, string(body))
			}
		})
	}
}

func TestDynamicWebsiteRoutingWithReferences(t *testing.T) {
	db := testDBForViews(t)
	withTempStorageForViews(t)

	headerFile := uploadHeaderForViews(t, "File", "header.html", `{{define "header"}}Header Component{{end}}`)
	headerVNode, err := p_filesystem.CreateVNode(db, "", false, headerFile, nil)
	if err != nil {
		t.Fatalf("failed to create header VNode: %v", err)
	}

	mainFile := uploadHeaderForViews(t, "File", "main.html", `{{template "header" .}}<div>Main Body</div>`)
	mainVNode, err := p_filesystem.CreateVNode(db, "", false, mainFile, nil)
	if err != nil {
		t.Fatalf("failed to create main VNode: %v", err)
	}

	route := DBRoute{
		Path:       "/referenced-page",
		PageID:     mainVNode.ID,
		References: []p_filesystem.VNode{*headerVNode},
		IsActive:   true,
	}
	if err := db.Create(&route).Error; err != nil {
		t.Fatalf("failed to create DBRoute with references: %v", err)
	}

	handler := buildHandler(t, db)

	req := httptest.NewRequest("GET", "/referenced-page", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", res.StatusCode, string(body))
	}
	if !strings.Contains(string(body), "Header Component") || !strings.Contains(string(body), "Main Body") {
		t.Errorf("expected body to contain referenced header and main body, got: %s", string(body))
	}
}
