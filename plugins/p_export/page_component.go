package p_export

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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

func (e exportPickerPage) Build(ctx context.Context) Node {
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
	if route, ok := lago.RegistryRoute.Get("export.DownloadRoute"); ok {
		action = route.Path
	}

	rows := Group{}
	for _, entry := range catalog.Entries {
		deps := "No auto-selected dependencies"
		if len(entry.ImmediateDeps) > 0 {
			deps = "Auto-selects: " + strings.Join(entry.ImmediateDeps, ", ")
		}
		description := fmt.Sprintf("%s columns", pluralize(len(entry.Columns), "1", fmt.Sprintf("%d", len(entry.Columns))))
		if entry.ModelName != "" && entry.ModelName != entry.Table {
			description = entry.ModelName + " | " + description
		}

		rows = append(rows,
			Div(
				Class("card bg-base-100 border border-base-300 shadow-sm"),
				Div(
					Class("card-body gap-3"),
					Div(
						Class("flex items-start justify-between gap-4"),
						Div(
							Class("space-y-1"),
							Div(Class("font-semibold"), Text(entry.Table)),
							Div(Class("text-sm text-base-content/70"), Text(description)),
							Div(Class("text-xs text-base-content/60"), Text(deps)),
						),
						Label(
							Class("label cursor-pointer gap-3"),
							Span(Class("label-text text-sm"), Text("Select")),
							Input(
								Type("checkbox"),
								Name("models"),
								Value(entry.Table),
								Class("checkbox checkbox-primary"),
								Attr("@change", "toggleRoot($event.target.value, $event.target.checked)"),
								Attr(":checked", fmt.Sprintf("isChecked(%q)", entry.Table)),
								Attr(":disabled", fmt.Sprintf("isAuto(%q)", entry.Table)),
							),
						),
					),
				),
			),
		)
	}

	if len(catalog.Entries) == 0 {
		rows = append(rows,
			Div(
				Class("alert"),
				Text("No registered models available in this deployment."),
			),
		)
	}

	return Div(
		Class("container max-w-6xl mx-auto mt-4"),
		Attr("x-data", exportPickerXData(string(depJSON))),
		Div(
			Class("mb-6"),
			H1(Class("text-2xl font-semibold"), Text("XLSX Export")),
			P(Class("text-sm text-base-content/70 mt-2"), Text("Select model roots. Frontend auto-selects dependencies. Backend recomputes same closure before export.")),
		),
		Div(
			Class("stats shadow mb-6 w-full"),
			Div(Class("stat"), Div(Class("stat-title"), Text("Models")), Div(Class("stat-value text-2xl"), Text(fmt.Sprintf("%d", len(catalog.Entries))))),
			Div(Class("stat"), Div(Class("stat-title"), Text("Root Selection")), Div(Class("stat-value text-2xl"), Span(Attr("x-text", "selectedRoots.length")))),
			Div(Class("stat"), Div(Class("stat-title"), Text("Effective Export")), Div(Class("stat-value text-2xl"), Span(Attr("x-text", "effective.length")))),
		),
		Form(
			Method("post"),
			Action(action),
			Class("space-y-6"),
			Attr("data-hx-boost", "false"),
			Div(Class("grid grid-cols-1 lg:grid-cols-2 gap-4"), rows),
			Div(
				Class("card bg-base-200 border border-base-300"),
				Div(
					Class("card-body gap-4"),
					Div(
						Class("text-sm"),
						Span(Class("font-semibold"), Text("Selected tables: ")),
						Span(Attr("x-text", `effective.length ? effective.join(", ") : "None"`)),
					),
					Div(
						Class("flex gap-3"),
						Button(Type("submit"), Class("btn btn-primary"), Text("Export XLSX")),
						Button(
							Type("button"),
							Class("btn btn-outline"),
							Attr("@click", "clearAll()"),
							Text("Clear"),
						),
					),
				),
			),
		),
	)
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
