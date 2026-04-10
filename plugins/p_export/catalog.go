package p_export

import (
	"fmt"
	"log/slog"
	"reflect"
	"sort"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ModelRelation struct {
	Name             string
	Type             string
	TargetTable      string
	JoinTable        string
	ForeignColumns   []string
	ReferenceColumns []string
	DisplayColumns   []string
}

type ModelCatalogEntry struct {
	Table         string
	ModelName     string
	Columns       []string
	PrimaryKeys   []string
	ImmediateDeps []string
	Relations     []ModelRelation
}

type JoinTableEntry struct {
	Table       string
	Columns     []string
	PrimaryKeys []string
}

type ExportCatalog struct {
	Entries    []ModelCatalogEntry
	JoinTables []JoinTableEntry

	byTable map[string]*ModelCatalogEntry
	byJoin  map[string]*JoinTableEntry
}

func BuildExportCatalog(db *gorm.DB) (ExportCatalog, error) {
	models := lago.RegistryModel.AllStable(registry.AlphabeticalByKey[any]{})
	entries := make([]ModelCatalogEntry, 0, len(*models))
	joinTables := map[string]JoinTableEntry{}

	for _, item := range *models {
		stmt := &gorm.Statement{DB: db}
		modelPtr := pointerForModel(item.Value)
		if err := stmt.Parse(modelPtr); err != nil {
			slog.Error("export: parse model schema", "table", item.Key, "error", err)
			return ExportCatalog{}, fmt.Errorf("parse model %q schema: %w", item.Key, err)
		}
		if stmt.Schema == nil {
			return ExportCatalog{}, fmt.Errorf("parse model %q schema: missing schema", item.Key)
		}

		entry := ModelCatalogEntry{
			Table:       stmt.Schema.Table,
			ModelName:   stmt.Schema.Name,
			Columns:     schemaColumns(stmt.Schema.Fields),
			PrimaryKeys: schemaColumns(stmt.Schema.PrimaryFields),
			Relations:   buildModelRelations(stmt.Schema),
		}
		entry.ImmediateDeps = immediateDependencies(entry.Relations, entry.Table)
		entries = append(entries, entry)

		for _, relation := range stmt.Schema.Relationships.Relations {
			if relation == nil || relation.JoinTable == nil || relation.JoinTable.Table == "" {
				continue
			}
			joinTable := JoinTableEntry{
				Table:       relation.JoinTable.Table,
				Columns:     schemaColumns(relation.JoinTable.Fields),
				PrimaryKeys: schemaColumns(relation.JoinTable.PrimaryFields),
			}
			joinTables[joinTable.Table] = joinTable
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Table < entries[j].Table
	})

	joins := make([]JoinTableEntry, 0, len(joinTables))
	for _, join := range joinTables {
		joins = append(joins, join)
	}
	sort.Slice(joins, func(i, j int) bool {
		return joins[i].Table < joins[j].Table
	})

	catalog := ExportCatalog{
		Entries:    entries,
		JoinTables: joins,
		byTable:    make(map[string]*ModelCatalogEntry, len(entries)),
		byJoin:     make(map[string]*JoinTableEntry, len(joins)),
	}
	for i := range catalog.Entries {
		entry := &catalog.Entries[i]
		catalog.byTable[entry.Table] = entry
	}
	for i := range catalog.JoinTables {
		entry := &catalog.JoinTables[i]
		catalog.byJoin[entry.Table] = entry
	}
	return catalog, nil
}

func (c ExportCatalog) Entry(table string) (*ModelCatalogEntry, bool) {
	entry, ok := c.byTable[table]
	return entry, ok
}

func (c ExportCatalog) Join(table string) (*JoinTableEntry, bool) {
	entry, ok := c.byJoin[table]
	return entry, ok
}

func buildModelRelations(s *schema.Schema) []ModelRelation {
	if s == nil {
		return nil
	}

	names := make([]string, 0, len(s.Relationships.Relations))
	for name := range s.Relationships.Relations {
		names = append(names, name)
	}
	sort.Strings(names)

	relations := make([]ModelRelation, 0, len(names))
	for _, name := range names {
		rel := s.Relationships.Relations[name]
		if rel == nil {
			continue
		}

		targetTable := ""
		if rel.FieldSchema != nil {
			targetTable = rel.FieldSchema.Table
		}
		joinTable := ""
		if rel.JoinTable != nil {
			joinTable = rel.JoinTable.Table
		}

		relation := ModelRelation{
			Name:             name,
			Type:             relationTypeString(rel.Type),
			TargetTable:      targetTable,
			JoinTable:        joinTable,
			ForeignColumns:   relationReferenceColumns(rel, true),
			ReferenceColumns: relationReferenceColumns(rel, false),
			DisplayColumns:   relationDisplayColumns(rel),
		}
		relations = append(relations, relation)
	}

	return relations
}

func relationTypeString(rt schema.RelationshipType) string {
	switch rt {
	case schema.HasOne:
		return "has_one"
	case schema.HasMany:
		return "has_many"
	case schema.BelongsTo:
		return "belongs_to"
	case schema.Many2Many:
		return "many_to_many"
	default:
		return fmt.Sprint(rt)
	}
}

func relationReferenceColumns(rel *schema.Relationship, foreign bool) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, ref := range rel.References {
		if ref == nil {
			continue
		}
		var field *schema.Field
		if foreign {
			field = ref.ForeignKey
		} else {
			field = ref.PrimaryKey
		}
		if field == nil || field.DBName == "" {
			continue
		}
		if _, ok := seen[field.DBName]; ok {
			continue
		}
		seen[field.DBName] = struct{}{}
		out = append(out, field.DBName)
	}
	sort.Strings(out)
	return out
}

func relationDisplayColumns(rel *schema.Relationship) []string {
	if rel == nil || rel.Type != schema.BelongsTo || rel.FieldSchema == nil {
		return nil
	}

	preferred := []string{"Name", "Title", "Email", "Username", "Code"}
	seen := map[string]struct{}{}
	out := []string{}

	for _, name := range preferred {
		field := rel.FieldSchema.LookUpField(name)
		if !isDisplayCandidate(field) {
			continue
		}
		seen[field.DBName] = struct{}{}
		out = append(out, field.DBName)
		if len(out) == 2 {
			return out
		}
	}

	for _, field := range rel.FieldSchema.Fields {
		if !isDisplayCandidate(field) {
			continue
		}
		if _, ok := seen[field.DBName]; ok {
			continue
		}
		seen[field.DBName] = struct{}{}
		out = append(out, field.DBName)
		if len(out) == 2 {
			break
		}
	}

	return out
}

func isDisplayCandidate(field *schema.Field) bool {
	if field == nil || field.DBName == "" {
		return false
	}
	if field.PrimaryKey {
		return false
	}
	if field.FieldType.Kind() != reflect.String {
		return false
	}
	return true
}

func schemaColumns(fields []*schema.Field) []string {
	columns := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == nil || field.DBName == "" {
			continue
		}
		columns = append(columns, field.DBName)
	}
	return columns
}

func immediateDependencies(relations []ModelRelation, sourceTable string) []string {
	seen := map[string]struct{}{}
	deps := []string{}
	for _, relation := range relations {
		if relation.TargetTable == "" || relation.TargetTable == sourceTable {
			continue
		}
		if relation.Type != "belongs_to" && relation.Type != "many_to_many" {
			continue
		}
		if _, ok := seen[relation.TargetTable]; ok {
			continue
		}
		seen[relation.TargetTable] = struct{}{}
		deps = append(deps, relation.TargetTable)
	}
	sort.Strings(deps)
	return deps
}

func pointerForModel(model any) any {
	value := reflect.ValueOf(model)
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.New(value.Type().Elem()).Interface()
		}
		return model
	}
	ptr := reflect.New(value.Type())
	ptr.Elem().Set(value)
	return ptr.Interface()
}
