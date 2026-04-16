package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ components.InputInterface = InputBulkAcademicRecordCourses{}

// InputBulkAcademicRecordCourses renders compulsory + optional course checkboxes for
// bulk submission creation; posts selected IDs as one JSON array in a hidden field.
type InputBulkAcademicRecordCourses struct {
	components.Page
	Label string
	Name  string // hidden field name for JSON array of course IDs
}

func (e InputBulkAcademicRecordCourses) GetKey() string     { return e.Key }
func (e InputBulkAcademicRecordCourses) GetRoles() []string { return e.Roles }

type bulkCourseCheckbox struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Section string `json:"section"`
	Checked bool   `json:"checked"`
}

func bulkAcademicRecordFromContext(ctx context.Context) (p_nirmancampus_academicrecords.AcademicRecord, bool) {
	rec, ok := ctx.Value(bulkAcademicRecordContextKey).(p_nirmancampus_academicrecords.AcademicRecord)
	return rec, ok && rec.ID != 0
}

func (e InputBulkAcademicRecordCourses) Build(ctx context.Context) Node {
	rec, ok := bulkAcademicRecordFromContext(ctx)
	if !ok {
		return Div(Class("text-sm text-error"), Text("Academic record not loaded."))
	}

	selected := map[uint]struct{}{}
	if inMap, ok := ctx.Value(getters.ContextKeyIn).(map[string]any); ok {
		switch raw := inMap[e.Name].(type) {
		case []uint:
			for _, id := range raw {
				selected[id] = struct{}{}
			}
		case []int:
			for _, id := range raw {
				if id > 0 {
					selected[uint(id)] = struct{}{}
				}
			}
		}
	}

	seen := map[uint]struct{}{}
	var items []bulkCourseCheckbox
	add := func(c p_nirmancampus_courses.Course, section string) {
		if c.ID == 0 {
			return
		}
		if _, dup := seen[c.ID]; dup {
			return
		}
		seen[c.ID] = struct{}{}
		_, sel := selected[c.ID]
		items = append(items, bulkCourseCheckbox{ID: c.ID, Name: c.Name, Section: section, Checked: sel})
	}
	for _, c := range rec.CompulsoryCourses {
		add(c, "compulsory")
	}
	for _, c := range rec.OptionalCourses {
		add(c, "optional")
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		slog.Error("InputBulkAcademicRecordCourses marshal failed", "error", err, "key", e.Key)
		itemsJSON = []byte("[]")
	}

	label := e.Label
	if label == "" {
		label = "Courses"
	}

	nameLit, err := json.Marshal(e.Name)
	if err != nil {
		slog.Error("InputBulkAcademicRecordCourses marshal name failed", "error", err, "key", e.Key)
		nameLit = []byte(`""`)
	}
	initJS := fmt.Sprintf(`
$el.closest('form').addEventListener('submit', (ev) => {
	const d = Alpine.$data($el);
	if (!d || !Array.isArray(d.items)) return;
	const ids = d.items.filter(item => item.checked).map(item => Number(item.id)).filter(id => id > 0);
	const h = $el.querySelector('input[type="hidden"][name=%s]');
	if (h) h.value = JSON.stringify(ids);
}, true);
`, string(nameLit))

	return Div(Class("my-1 flex flex-col gap-3"),
		Label(Class("label text-sm font-bold"), Text(label)),
		Div(
			Class("flex flex-col gap-2"),
			Attr("x-data", fmt.Sprintf(`{ items: %s }`, string(itemsJSON))),
			Attr("x-init", initJS),
			Div(Class("text-xs font-semibold uppercase opacity-70"), Text("Compulsory")),
			Template(
				Attr("x-for", "item in items.filter(i => i.section === 'compulsory')"),
				Attr(":key", "'c-'+item.id"),
				Label(Class("label cursor-pointer flex flex-row items-center gap-2 text-sm"),
					Input(
						Type("checkbox"),
						Class("checkbox"),
						Attr("x-model", "item.checked"),
					),
					Span(Attr("x-text", "item.name")),
				),
			),
			Div(Class("text-xs font-semibold uppercase opacity-70 mt-2"), Text("Optional")),
			Template(
				Attr("x-for", "item in items.filter(i => i.section === 'optional')"),
				Attr(":key", "'o-'+item.id"),
				Label(Class("label cursor-pointer flex flex-row items-center gap-2 text-sm"),
					Input(
						Type("checkbox"),
						Class("checkbox"),
						Attr("x-model", "item.checked"),
					),
					Span(Attr("x-text", "item.name")),
				),
			),
			Input(Type("hidden"), Name(e.Name)),
		),
	)
}

func (e InputBulkAcademicRecordCourses) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		return []uint(nil), nil
	}
	raw := strings.TrimSpace(vals[0])
	var ids []uint
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, fmt.Errorf("invalid course selection: %w", err)
	}
	out := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (e InputBulkAcademicRecordCourses) GetName() string {
	return e.Name
}
