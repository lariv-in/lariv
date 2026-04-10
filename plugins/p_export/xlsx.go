package p_export

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func BuildWorkbook(db *gorm.DB, catalog ExportCatalog, selection ExpandedSelection) (*excelize.File, error) {
	if len(selection.Tables) == 0 {
		return nil, fmt.Errorf("no tables selected for export")
	}

	workbook := excelize.NewFile()
	usedNames := map[string]struct{}{}
	tableSheets := assignSheetNames(selection.Tables, usedNames)
	joinSheets := assignSheetNames(selection.JoinTables, usedNames)
	relationsSheet := uniqueSheetName("_relations", usedNames)

	firstSheet := true
	for _, table := range selection.Tables {
		entry, ok := catalog.Entry(table)
		if !ok {
			return nil, fmt.Errorf("missing model catalog entry for %q", table)
		}
		rows, err := fetchTableRows(db, table, entry.PrimaryKeys)
		if err != nil {
			return nil, err
		}
		headers, enrichedRows, err := enrichModelRows(db, catalog, entry, rows)
		if err != nil {
			return nil, err
		}
		if err := writeSheet(workbook, tableSheets[table], headers, enrichedRows, firstSheet); err != nil {
			return nil, err
		}
		firstSheet = false
	}

	for _, table := range selection.JoinTables {
		entry, ok := catalog.Join(table)
		if !ok {
			continue
		}
		rows, err := fetchTableRows(db, table, entry.PrimaryKeys)
		if err != nil {
			return nil, err
		}
		if err := writeSheet(workbook, joinSheets[table], entry.Columns, rows, firstSheet); err != nil {
			return nil, err
		}
		firstSheet = false
	}

	relationRows := buildRelationRows(catalog, selection, tableSheets, joinSheets)
	if err := writeSheet(workbook, relationsSheet, relationHeaders(), relationRows, firstSheet); err != nil {
		return nil, err
	}

	return workbook, nil
}

func buildRelationRows(catalog ExportCatalog, selection ExpandedSelection, tableSheets, joinSheets map[string]string) []map[string]any {
	includedTables := map[string]struct{}{}
	for _, table := range selection.Tables {
		includedTables[table] = struct{}{}
	}
	includedJoins := map[string]struct{}{}
	for _, table := range selection.JoinTables {
		includedJoins[table] = struct{}{}
	}

	rows := []map[string]any{}
	for _, table := range selection.Tables {
		entry, ok := catalog.Entry(table)
		if !ok {
			continue
		}
		for _, relation := range entry.Relations {
			if relation.TargetTable != "" {
				if _, ok := includedTables[relation.TargetTable]; !ok {
					continue
				}
			}
			if relation.JoinTable != "" {
				if _, ok := includedJoins[relation.JoinTable]; !ok {
					continue
				}
			}
			rows = append(rows, map[string]any{
				"source_table":      entry.Table,
				"source_sheet":      tableSheets[entry.Table],
				"relation_name":     relation.Name,
				"relation_type":     relation.Type,
				"target_table":      relation.TargetTable,
				"target_sheet":      tableSheets[relation.TargetTable],
				"join_table":        relation.JoinTable,
				"join_sheet":        joinSheets[relation.JoinTable],
				"foreign_columns":   strings.Join(relation.ForeignColumns, ","),
				"reference_columns": strings.Join(relation.ReferenceColumns, ","),
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		left := fmt.Sprintf("%s|%s|%s", rows[i]["source_table"], rows[i]["relation_name"], rows[i]["target_table"])
		right := fmt.Sprintf("%s|%s|%s", rows[j]["source_table"], rows[j]["relation_name"], rows[j]["target_table"])
		return left < right
	})

	return rows
}

func relationHeaders() []string {
	return []string{
		"source_table",
		"source_sheet",
		"relation_name",
		"relation_type",
		"target_table",
		"target_sheet",
		"join_table",
		"join_sheet",
		"foreign_columns",
		"reference_columns",
	}
}

func fetchTableRows(db *gorm.DB, table string, primaryKeys []string) ([]map[string]any, error) {
	rows := []map[string]any{}
	query := db.Table(table)
	for _, column := range primaryKeys {
		query = query.Order(fmt.Sprintf(`"%s" ASC`, column))
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load rows for %q: %w", table, err)
	}
	if len(primaryKeys) == 0 {
		sortRows(rows)
	}
	return rows, nil
}

func enrichModelRows(db *gorm.DB, catalog ExportCatalog, entry *ModelCatalogEntry, rows []map[string]any) ([]string, []map[string]any, error) {
	if entry == nil {
		return nil, nil, fmt.Errorf("missing model catalog entry")
	}

	headers := append([]string(nil), entry.Columns...)
	enrichedRows := copyRows(rows)

	lookups := map[string]relationLookup{}
	for _, relation := range entry.Relations {
		if !supportsDisplayColumns(relation) {
			continue
		}
		if _, ok := catalog.Entry(relation.TargetTable); !ok {
			continue
		}

		prefix := relationColumnPrefix(relation)
		for _, column := range relation.DisplayColumns {
			headers = append(headers, prefix+"__"+column)
		}

		lookup, ok := lookups[relation.TargetTable]
		if !ok {
			var err error
			lookup, err = buildRelationLookup(db, relation)
			if err != nil {
				return nil, nil, err
			}
			lookups[relation.TargetTable] = lookup
		}

		for _, row := range enrichedRows {
			values := lookup.values[buildCompositeKeyFromRow(row, relation.ForeignColumns)]
			for _, column := range relation.DisplayColumns {
				key := prefix + "__" + column
				row[key] = values[column]
			}
		}
	}

	return headers, enrichedRows, nil
}

type relationLookup struct {
	values map[string]map[string]any
}

func buildRelationLookup(db *gorm.DB, relation ModelRelation) (relationLookup, error) {
	selectColumns := append([]string(nil), relation.ReferenceColumns...)
	for _, column := range relation.DisplayColumns {
		if !slicesContains(selectColumns, column) {
			selectColumns = append(selectColumns, column)
		}
	}

	rows, err := fetchSelectedRows(db, relation.TargetTable, selectColumns, relation.ReferenceColumns)
	if err != nil {
		return relationLookup{}, err
	}

	lookup := relationLookup{
		values: make(map[string]map[string]any, len(rows)),
	}
	for _, row := range rows {
		lookup.values[buildCompositeKeyFromRow(row, relation.ReferenceColumns)] = row
	}
	return lookup, nil
}

func fetchSelectedRows(db *gorm.DB, table string, columns, orderColumns []string) ([]map[string]any, error) {
	rows := []map[string]any{}
	query := db.Table(table).Select(strings.Join(columns, ", "))
	for _, column := range orderColumns {
		query = query.Order(fmt.Sprintf(`"%s" ASC`, column))
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load relation rows for %q: %w", table, err)
	}
	return rows, nil
}

func supportsDisplayColumns(relation ModelRelation) bool {
	return relation.Type == "belongs_to" &&
		relation.TargetTable != "" &&
		len(relation.ForeignColumns) > 0 &&
		len(relation.ReferenceColumns) > 0 &&
		len(relation.ForeignColumns) == len(relation.ReferenceColumns) &&
		len(relation.DisplayColumns) > 0
}

func relationColumnPrefix(relation ModelRelation) string {
	if len(relation.ForeignColumns) == 1 && strings.HasSuffix(relation.ForeignColumns[0], "_id") {
		return strings.TrimSuffix(relation.ForeignColumns[0], "_id")
	}
	name := toSnakeCase(relation.Name)
	if name != "" {
		return name
	}
	return relation.TargetTable
}

func buildCompositeKeyFromRow(row map[string]any, columns []string) string {
	parts := make([]string, len(columns))
	for i, column := range columns {
		parts[i] = stringifyKeyValue(row[column])
	}
	return strings.Join(parts, "|")
}

func stringifyKeyValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	case fmt.Stringer:
		return typed.String()
	case json.Number:
		return typed.String()
	case int:
		return strconv.Itoa(typed)
	case int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(typed).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(typed).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(typed).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(value)
	}
}

func copyRows(rows []map[string]any) []map[string]any {
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		copyRow := make(map[string]any, len(row))
		for key, value := range row {
			copyRow[key] = value
		}
		out[i] = copyRow
	}
	return out
}

func slicesContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') {
				out = append(out, '_')
			}
		}
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		out = append(out, r)
	}
	return string(out)
}

func sortRows(rows []map[string]any) {
	sort.Slice(rows, func(i, j int) bool {
		return stringifyRow(rows[i]) < stringifyRow(rows[j])
	})
}

func stringifyRow(row map[string]any) string {
	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, row[key]))
	}
	return strings.Join(parts, "|")
}

func writeSheet(workbook *excelize.File, sheet string, headers []string, rows []map[string]any, useDefaultSheet bool) error {
	if useDefaultSheet {
		if err := workbook.SetSheetName("Sheet1", sheet); err != nil {
			return err
		}
	} else {
		workbook.NewSheet(sheet)
	}

	headerValues := make([]any, len(headers))
	for i, header := range headers {
		headerValues[i] = header
	}
	if err := workbook.SetSheetRow(sheet, "A1", &headerValues); err != nil {
		return err
	}

	headerStyle, err := workbook.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err == nil {
		lastColumn, cellErr := excelize.ColumnNumberToName(len(headers))
		if cellErr == nil {
			_ = workbook.SetCellStyle(sheet, "A1", lastColumn+"1", headerStyle)
		}
	}

	for i, row := range rows {
		values := make([]any, len(headers))
		for col, header := range headers {
			values[col] = normalizeCellValue(row[header])
		}
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			return err
		}
		if err := workbook.SetSheetRow(sheet, cell, &values); err != nil {
			return err
		}
	}

	_ = workbook.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
	_ = workbook.AutoFilter(sheet, fmt.Sprintf("A1:%s1", lastColumnName(len(headers))), nil)
	for i := range headers {
		columnName, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		_ = workbook.SetColWidth(sheet, columnName, columnName, 18)
	}

	return nil
}

func normalizeCellValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return ""
	case time.Time:
		return typed.Format(time.RFC3339Nano)
	case *time.Time:
		if typed == nil {
			return ""
		}
		return typed.Format(time.RFC3339Nano)
	case []byte:
		if utf8.Valid(typed) {
			return string(typed)
		}
		return base64.StdEncoding.EncodeToString(typed)
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return typed
	default:
		data, err := json.Marshal(typed)
		if err == nil {
			return string(data)
		}
		return fmt.Sprint(typed)
	}
}

func assignSheetNames(names []string, used map[string]struct{}) map[string]string {
	out := make(map[string]string, len(names))
	for _, name := range names {
		out[name] = uniqueSheetName(name, used)
	}
	return out
}

func uniqueSheetName(name string, used map[string]struct{}) string {
	clean := sanitizeSheetName(name)
	if clean == "" {
		clean = "Sheet"
	}
	if len(clean) > 31 {
		clean = clean[:31]
	}
	candidate := clean
	for i := 2; ; i++ {
		if _, exists := used[candidate]; !exists {
			used[candidate] = struct{}{}
			return candidate
		}
		suffix := fmt.Sprintf("_%d", i)
		base := clean
		if len(base)+len(suffix) > 31 {
			base = base[:31-len(suffix)]
		}
		candidate = base + suffix
	}
}

func sanitizeSheetName(name string) string {
	replacer := strings.NewReplacer(
		":", "_",
		"\\", "_",
		"/", "_",
		"?", "_",
		"*", "_",
		"[", "_",
		"]", "_",
	)
	name = replacer.Replace(strings.TrimSpace(name))
	name = strings.Trim(name, "'")
	return name
}

func lastColumnName(count int) string {
	if count <= 0 {
		return "A"
	}
	name, err := excelize.ColumnNumberToName(count)
	if err != nil {
		return "A"
	}
	return name
}
