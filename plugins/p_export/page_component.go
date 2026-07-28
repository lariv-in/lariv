package p_export

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
)

type exportPickerPage struct {
	components.Page
}

func (e exportPickerPage) GetKey() string {
	return e.Key
}

func (e exportPickerPage) GetRoles() []string {
	return e.Roles
}

type exportPickerRow struct {
	Table        string
	Description  string
	Deps         string
	CheckedExpr  string
	DisabledExpr string
}

func (e exportPickerPage) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	catalog, _ := ctx.Value(exportCatalogContextKey).(ExportCatalog)
	dependencyMap := make(map[string][]string, len(catalog.Entries))
	for _, entry := range catalog.Entries {
		dependencyMap[entry.Table] = entry.ImmediateDeps
	}

	depJSON, err := json.Marshal(dependencyMap)
	if err != nil {
		slog.Error("export: marshal dependency map", "error", err)
		depJSON = []byte("{}")
	}

	action := "#"
	if app, ok := lariv.AppFromContext(ctx); ok {
		if route, found := app.Routes.Get("export.DownloadRoute"); found {
			action = route.Path
		}
	}

	rows := make([]exportPickerRow, 0, len(catalog.Entries))
	for _, entry := range catalog.Entries {
		deps := "No auto-selected dependencies"
		if len(entry.ImmediateDeps) > 0 {
			deps = "Auto-selects: " + strings.Join(entry.ImmediateDeps, ", ")
		}
		description := fmt.Sprintf("%s columns", pluralize(len(entry.Columns), "1", fmt.Sprintf("%d", len(entry.Columns))))
		if entry.ModelName != "" && entry.ModelName != entry.Table {
			description = entry.ModelName + " | " + description
		}

		rows = append(rows, exportPickerRow{
			Table:        entry.Table,
			Description:  description,
			Deps:         deps,
			CheckedExpr:  fmt.Sprintf("isChecked(%q)", entry.Table),
			DisabledExpr: fmt.Sprintf("isAuto(%q)", entry.Table),
		})
	}

	return execute(w, "export_picker", struct {
		XData      string
		ModelCount int
		Action     string
		Rows       []exportPickerRow
	}{
		XData:      exportPickerXData(string(depJSON)),
		ModelCount: len(catalog.Entries),
		Action:     action,
		Rows:       rows,
	})
}

func exportPickerXData(depJSON string) string {
	return fmt.Sprintf(`{
		deps: %s,
		selectedRoots: [],
		effective: [],
		init() { this.recompute(); },
		toggleRoot(table, checked) {
			if (checked) {
				if (!this.selectedRoots.includes(table)) this.selectedRoots.push(table);
			} else {
				this.selectedRoots = this.selectedRoots.filter((item) => item !== table);
			}
			this.recompute();
		},
		recompute() {
			const effective = new Set(this.selectedRoots);
			let changed = true;
			while (changed) {
				changed = false;
				for (const table of Array.from(effective)) {
					for (const dep of (this.deps[table] || [])) {
						if (!effective.has(dep)) {
							effective.add(dep);
							changed = true;
						}
					}
				}
			}
			this.effective = Array.from(effective).sort();
		},
		isChecked(table) {
			return this.effective.includes(table);
		},
		isAuto(table) {
			return this.isChecked(table) && !this.selectedRoots.includes(table);
		},
		clearAll() {
			this.selectedRoots = [];
			this.recompute();
		}
	}`, depJSON)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
