package p_filesystem

import (
	"testing"
)

func TestDatabaseFilesystem(t *testing.T) {
	db := testDB(t)
	withTempStorage(t)

	// Create directories
	dir1, err := CreateVNode(db, "dir1", true, nil, nil)
	if err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}

	_, err = CreateVNode(db, "dir2", true, nil, dir1)
	if err != nil {
		t.Fatalf("failed to create dir2: %v", err)
	}

	// Create file in dir1
	fileHeader := uploadHeader(t, "File", "test.txt", "hello database filesystem")
	_, err = CreateVNode(db, "", false, fileHeader, dir1)
	if err != nil {
		t.Fatalf("failed to create test.txt: %v", err)
	}

	dfs := NewDatabaseFilesystem(db)

	// Test ReadDir of root
	entries, err := dfs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir root failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "dir1" {
		t.Fatalf("expected exactly dir1 in root, got: %v", entries)
	}
	if !entries[0].IsDir() {
		t.Fatalf("expected dir1 to be a directory")
	}

	// Test ReadDir of dir1
	entries, err = dfs.ReadDir("dir1")
	if err != nil {
		t.Fatalf("ReadDir dir1 failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries in dir1, got %d", len(entries))
	}
	// order is directory DESC, name ASC. So dir2 is first, then test.txt.
	if entries[0].Name() != "dir2" || !entries[0].IsDir() {
		t.Errorf("expected first entry to be dir2, got %v", entries[0])
	}
	if entries[1].Name() != "test.txt" || entries[1].IsDir() {
		t.Errorf("expected second entry to be test.txt, got %v", entries[1])
	}

	// Test ReadFile
	data, err := dfs.ReadFile("dir1/test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "hello database filesystem" {
		t.Fatalf("unexpected content: %q", string(data))
	}

	// Test opening a missing file
	_, err = dfs.Open("dir1/missing.txt")
	if err == nil {
		t.Fatalf("expected error when opening missing file")
	}

	// Test opening an invalid path
	_, err = dfs.Open("dir1/../dir1")
	if err == nil {
		t.Fatalf("expected error when opening invalid path")
	}

	// Test Open and Read on directory
	dirFile, err := dfs.Open("dir1")
	if err != nil {
		t.Fatalf("failed to open dir1: %v", err)
	}
	defer dirFile.Close()

	buf := make([]byte, 10)
	_, err = dirFile.Read(buf)
	if err == nil {
		t.Fatalf("expected error reading from a directory file")
	}
}
