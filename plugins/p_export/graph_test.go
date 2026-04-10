package p_export

import (
	"reflect"
	"testing"
)

func TestExpandSelectionAddsTransitiveDepsAndJoinTables(t *testing.T) {
	catalog := testCatalog(
		ModelCatalogEntry{
			Table:         "groups",
			ImmediateDeps: []string{"users"},
			Relations: []ModelRelation{
				{Type: "many_to_many", TargetTable: "users", JoinTable: "group_users"},
			},
		},
		ModelCatalogEntry{
			Table:         "users",
			ImmediateDeps: []string{"roles"},
			Relations: []ModelRelation{
				{Type: "belongs_to", TargetTable: "roles"},
			},
		},
		ModelCatalogEntry{
			Table: "roles",
		},
	)

	got, err := ExpandSelection(catalog, []string{"groups"})
	if err != nil {
		t.Fatalf("ExpandSelection error = %v", err)
	}

	wantTables := []string{"groups", "roles", "users"}
	if !reflect.DeepEqual(got.Tables, wantTables) {
		t.Fatalf("Tables = %v, want %v", got.Tables, wantTables)
	}

	wantJoins := []string{"group_users"}
	if !reflect.DeepEqual(got.JoinTables, wantJoins) {
		t.Fatalf("JoinTables = %v, want %v", got.JoinTables, wantJoins)
	}
}

func TestExpandSelectionRejectsUnknownTable(t *testing.T) {
	catalog := testCatalog(ModelCatalogEntry{Table: "users"})

	_, err := ExpandSelection(catalog, []string{"missing"})
	if err == nil {
		t.Fatal("expected error for unknown table")
	}
}

func testCatalog(entries ...ModelCatalogEntry) ExportCatalog {
	catalog := ExportCatalog{
		Entries: entries,
		byTable: make(map[string]*ModelCatalogEntry, len(entries)),
		byJoin:  map[string]*JoinTableEntry{},
	}
	for i := range catalog.Entries {
		entry := &catalog.Entries[i]
		catalog.byTable[entry.Table] = entry
	}
	return catalog
}
