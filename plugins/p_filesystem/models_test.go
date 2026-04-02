package p_filesystem

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	if err := db.AutoMigrate(&VNode{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS filesystem_nodes_parent_name_dir_uidx ON filesystem_nodes (COALESCE(parent_id, 0), name, is_directory)",
	).Error; err != nil {
		t.Fatalf("creating unique index failed: %v", err)
	}
	return db
}

func withTempStorage(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	previous := Store
	Store = &LocalFilestore{BaseDir: dir}
	t.Cleanup(func() {
		Store = previous
	})
	return dir
}

func uploadHeader(t *testing.T, fieldName, fileName, body string) *multipart.FileHeader {
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

func TestCreateVNodeAndOpenDownload(t *testing.T) {
	db := testDB(t)
	withTempStorage(t)

	dir, err := CreateVNode(db, "docs", true, nil, nil)
	if err != nil {
		t.Fatalf("CreateVNode directory failed: %v", err)
	}

	file := uploadHeader(t, "File", "report.txt", "hello world")
	node, err := CreateVNode(db, "", false, file, dir)
	if err != nil {
		t.Fatalf("CreateVNode file failed: %v", err)
	}
	if node.Name != "report.txt" {
		t.Fatalf("expected file name to default from upload, got %q", node.Name)
	}

	download, err := node.OpenDownload()
	if err != nil {
		t.Fatalf("OpenDownload failed: %v", err)
	}
	defer download.Reader.Close()

	data, err := io.ReadAll(download.Reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected download contents to match upload, got %q", string(data))
	}
}

func TestMoveToNodeRejectsDescendant(t *testing.T) {
	db := testDB(t)
	withTempStorage(t)

	parent, err := CreateVNode(db, "parent", true, nil, nil)
	if err != nil {
		t.Fatalf("CreateVNode parent failed: %v", err)
	}
	child, err := CreateVNode(db, "child", true, nil, parent)
	if err != nil {
		t.Fatalf("CreateVNode child failed: %v", err)
	}

	if err := parent.MoveToNode(db, child); err == nil {
		t.Fatalf("expected move into descendant to fail")
	}
}

func TestDeleteTreeRemovesStoredFiles(t *testing.T) {
	db := testDB(t)
	storageDir := withTempStorage(t)

	parent, err := CreateVNode(db, "parent", true, nil, nil)
	if err != nil {
		t.Fatalf("CreateVNode parent failed: %v", err)
	}
	file := uploadHeader(t, "File", "child.txt", "delete me")
	node, err := CreateVNode(db, "", false, file, parent)
	if err != nil {
		t.Fatalf("CreateVNode child file failed: %v", err)
	}

	if _, err := os.Stat(node.FilePath); err != nil {
		t.Fatalf("expected stored file to exist before delete: %v", err)
	}

	if err := parent.DeleteTree(db); err != nil {
		t.Fatalf("DeleteTree failed: %v", err)
	}

	if _, err := os.Stat(node.FilePath); !os.IsNotExist(err) {
		t.Fatalf("expected stored file to be removed, stat err = %v", err)
	}

	var count int64
	if err := db.Model(&VNode{}).Count(&count).Error; err != nil {
		t.Fatalf("counting nodes failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected all nodes deleted, got count %d", count)
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected storage dir to be empty after delete, found %d entries", len(entries))
	}
}
