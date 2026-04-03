package views

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

// applyListViewSorts applies ?sort= values to a list query. Dotted paths (e.g. User.Name)
// require BelongsTo or HasOne steps only; GORM Joins are applied with deduplicated paths,
// then ORDER BY uses join aliases (User, User__Company, …) matching GORM's generator.
// Single-segment sorts use the base model column; when any join was added, base columns
// are qualified with the base table name so ORDER BY stays unambiguous.
//
// BelongsTo/HasOne joins do not multiply base rows, so Count + pagination stay correct.
func applyListViewSorts[T any](query gorm.ChainInterface[T], root *schema.Schema, sortValues []string) gorm.ChainInterface[T] {
	if root == nil {
		return query
	}

	type pendingOrder struct {
		dotted    bool
		dir       string
		baseField *schema.Field
		alias     string
		leafDB    string
	}

	var joinOrder []string
	joinSeen := map[string]struct{}{}
	var pending []pendingOrder

	for _, raw := range sortValues {
		ident, dir := parseSortIdentAndDir(raw)
		if ident == "" {
			continue
		}
		segs := splitSortPathSegments(ident)
		if len(segs) == 0 {
			continue
		}
		if len(segs) == 1 {
			f := root.LookUpField(segs[0])
			if f == nil || f.DBName == "" || !f.Readable {
				continue
			}
			pending = append(pending, pendingOrder{dotted: false, dir: dir, baseField: f})
			continue
		}

		relNames := segs[:len(segs)-1]
		leafName := segs[len(segs)-1]
		joinPath := strings.Join(relNames, ".")

		current := root
		for _, rn := range relNames {
			rn = strings.TrimSpace(rn)
			if rn == "" {
				current = nil
				break
			}
			rel, ok := current.Relationships.Relations[rn]
			if !ok || rel == nil || rel.FieldSchema == nil {
				current = nil
				break
			}
			if rel.Type != schema.BelongsTo && rel.Type != schema.HasOne {
				current = nil
				break
			}
			current = rel.FieldSchema
		}
		if current == nil {
			continue
		}
		leafField := current.LookUpField(leafName)
		if leafField == nil || leafField.DBName == "" || !leafField.Readable {
			continue
		}
		if _, ok := joinSeen[joinPath]; !ok {
			joinSeen[joinPath] = struct{}{}
			joinOrder = append(joinOrder, joinPath)
		}
		alias := utils.JoinNestedRelationNames(relNames)
		pending = append(pending, pendingOrder{dotted: true, dir: dir, alias: alias, leafDB: leafField.DBName})
	}

	for _, jp := range joinOrder {
		query = query.Joins(clause.Has(jp), nil)
	}

	anyJoin := len(joinOrder) > 0
	baseTable := root.Table
	for _, p := range pending {
		if p.dotted {
			query = query.Order(quoteTableColumn(p.alias, p.leafDB) + p.dir)
			continue
		}
		if anyJoin {
			query = query.Order(quoteTableColumn(baseTable, p.baseField.DBName) + p.dir)
		} else {
			query = query.Order(p.baseField.DBName + p.dir)
		}
	}

	return query
}

func parseSortIdentAndDir(raw string) (ident, dir string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", ""
	}
	colTokens := parts
	if n := len(parts); n >= 2 {
		last := strings.ToUpper(parts[n-1])
		if last == "ASC" || last == "DESC" {
			dir = " " + last
			colTokens = parts[:n-1]
		}
	}
	if len(colTokens) == 0 {
		return "", dir
	}
	return strings.Join(colTokens, " "), dir
}

func splitSortPathSegments(ident string) []string {
	chunks := strings.Split(ident, ".")
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c == "" {
			return nil
		}
		out = append(out, c)
	}
	return out
}

func quoteTableColumn(tableOrAlias, column string) string {
	return fmt.Sprintf(`"%s"."%s"`, escapeSQLIdent(tableOrAlias), escapeSQLIdent(column))
}

func escapeSQLIdent(s string) string {
	return strings.ReplaceAll(s, `"`, `""`)
}
