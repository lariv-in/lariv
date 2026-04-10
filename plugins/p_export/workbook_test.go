package p_export

import (
	"slices"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildWorkbookCreatesModelJoinAndRelationSheets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE roles (id integer primary key, name text)`,
		`CREATE TABLE users (id integer primary key, role_id integer, name text)`,
		`CREATE TABLE groups (id integer primary key, name text)`,
		`CREATE TABLE group_users (group_id integer, user_id integer)`,
		`INSERT INTO roles (id, name) VALUES (1, 'admin')`,
		`INSERT INTO users (id, role_id, name) VALUES (10, 1, 'sandy')`,
		`INSERT INTO users (id, role_id, name) VALUES (11, 99, 'orphan')`,
		`INSERT INTO groups (id, name) VALUES (20, 'editors')`,
		`INSERT INTO group_users (group_id, user_id) VALUES (20, 10)`,
		`INSERT INTO group_users (group_id, user_id) VALUES (20, 11)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("exec %q: %v", statement, err)
		}
	}

	catalog := testCatalog(
		ModelCatalogEntry{
			Table:       "roles",
			Columns:     []string{"id", "name"},
			PrimaryKeys: []string{"id"},
		},
		ModelCatalogEntry{
			Table:         "users",
			Columns:       []string{"id", "role_id", "name"},
			PrimaryKeys:   []string{"id"},
			ImmediateDeps: []string{"roles"},
			Relations: []ModelRelation{
				{Name: "Role", Type: "belongs_to", TargetTable: "roles", ForeignColumns: []string{"role_id"}, ReferenceColumns: []string{"id"}, DisplayColumns: []string{"name"}},
			},
		},
		ModelCatalogEntry{
			Table:         "groups",
			Columns:       []string{"id", "name"},
			PrimaryKeys:   []string{"id"},
			ImmediateDeps: []string{"users"},
			Relations: []ModelRelation{
				{Name: "Users", Type: "many_to_many", TargetTable: "users", JoinTable: "group_users", ForeignColumns: []string{"group_id", "user_id"}, ReferenceColumns: []string{"id"}},
			},
		},
	)
	catalog.JoinTables = []JoinTableEntry{
		{Table: "group_users", Columns: []string{"group_id", "user_id"}},
	}
	catalog.byJoin["group_users"] = &catalog.JoinTables[0]

	selection, err := ExpandSelection(catalog, []string{"groups"})
	if err != nil {
		t.Fatalf("ExpandSelection: %v", err)
	}

	workbook, err := BuildWorkbook(db, catalog, selection)
	if err != nil {
		t.Fatalf("BuildWorkbook: %v", err)
	}
	defer workbook.Close()

	sheets := workbook.GetSheetList()
	for _, sheet := range []string{"groups", "users", "roles", "group_users", "_relations"} {
		if !slices.Contains(sheets, sheet) {
			t.Fatalf("missing sheet %q in %v", sheet, sheets)
		}
	}

	relationRows, err := workbook.GetRows("_relations")
	if err != nil {
		t.Fatalf("GetRows(_relations): %v", err)
	}
	if len(relationRows) < 2 {
		t.Fatalf("expected relation metadata rows, got %v", relationRows)
	}
	if relationRows[0][0] != "source_table" || relationRows[1][0] != "groups" {
		t.Fatalf("unexpected relation rows: %v", relationRows)
	}

	userRows, err := workbook.GetRows("users")
	if err != nil {
		t.Fatalf("GetRows(users): %v", err)
	}
	if len(userRows) != 3 {
		t.Fatalf("expected header plus 2 user rows, got %v", userRows)
	}
	expectedHeader := []string{"id", "role_id", "name", "role__name"}
	for i, value := range expectedHeader {
		if userRows[0][i] != value {
			t.Fatalf("users header[%d] = %q, want %q; full=%v", i, userRows[0][i], value, userRows[0])
		}
	}
	if userRows[1][3] != "admin" {
		t.Fatalf("expected resolved role display column, got %v", userRows[1])
	}
	if len(userRows[2]) != 3 {
		t.Fatalf("expected orphan row to omit blank trailing helper cell, got %v", userRows[2])
	}
}
