package p_export

import (
	"fmt"
	"sort"
)

type ExpandedSelection struct {
	Roots      []string
	Tables     []string
	JoinTables []string
}

func ExpandSelection(catalog ExportCatalog, roots []string) (ExpandedSelection, error) {
	normalizedRoots := normalizeSelection(roots)
	if len(normalizedRoots) == 0 {
		return ExpandedSelection{}, fmt.Errorf("select at least one model")
	}

	included := map[string]struct{}{}
	queue := make([]string, 0, len(normalizedRoots))
	for _, table := range normalizedRoots {
		if _, ok := catalog.Entry(table); !ok {
			return ExpandedSelection{}, fmt.Errorf("unknown model table %q", table)
		}
		if _, ok := included[table]; ok {
			continue
		}
		included[table] = struct{}{}
		queue = append(queue, table)
	}

	joinTables := map[string]struct{}{}
	for len(queue) > 0 {
		table := queue[0]
		queue = queue[1:]

		entry, ok := catalog.Entry(table)
		if !ok {
			return ExpandedSelection{}, fmt.Errorf("missing catalog entry for %q", table)
		}

		for _, dep := range entry.ImmediateDeps {
			if _, ok := included[dep]; ok {
				continue
			}
			if _, ok := catalog.Entry(dep); !ok {
				continue
			}
			included[dep] = struct{}{}
			queue = append(queue, dep)
		}

		for _, relation := range entry.Relations {
			if relation.Type == "many_to_many" && relation.JoinTable != "" {
				joinTables[relation.JoinTable] = struct{}{}
			}
		}
	}

	tables := make([]string, 0, len(included))
	for table := range included {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	joins := make([]string, 0, len(joinTables))
	for table := range joinTables {
		joins = append(joins, table)
	}
	sort.Strings(joins)

	return ExpandedSelection{
		Roots:      normalizedRoots,
		Tables:     tables,
		JoinTables: joins,
	}, nil
}

func normalizeSelection(tables []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tables))
	for _, table := range tables {
		if table == "" {
			continue
		}
		if _, ok := seen[table]; ok {
			continue
		}
		seen[table] = struct{}{}
		out = append(out, table)
	}
	sort.Strings(out)
	return out
}
