package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ components.InputInterface = InputBulkSubmissionMarks{}

// InputBulkSubmissionMarks lists one number input per assignment submission; posts JSON array to hidden field.
type InputBulkSubmissionMarks struct {
	components.Page
	Label string
	Name  string
}

type bulkMarksItem struct {
	ID              uint   `json:"id"`
	AssignmentTitle string `json:"assignmentTitle"`
	CourseName      string `json:"courseName"`
	MaxMarks        int    `json:"maxMarks"`
	Marks           int    `json:"marks"`
}

func (e InputBulkSubmissionMarks) GetKey() string     { return e.Key }
func (e InputBulkSubmissionMarks) GetRoles() []string { return e.Roles }

func (e InputBulkSubmissionMarks) Build(ctx context.Context) Node {
	subs, ok := ctx.Value(bulkAddMarksSubmissionsKey).([]AssignmentSubmission)
	if !ok {
		return Div(Class("text-sm text-error"), Text("Submissions not loaded."))
	}

	var items []bulkMarksItem
	for _, s := range subs {
		cname := ""
		if s.Course.ID != 0 {
			cname = s.Course.Name
		}
		marks := s.Marks
		items = append(items, bulkMarksItem{
			ID:              s.ID,
			AssignmentTitle: s.AssignmentTitle,
			CourseName:      cname,
			MaxMarks:        s.MaxMarks,
			Marks:           marks,
		})
	}

	if inMap, ok2 := ctx.Value(getters.ContextKeyIn).(map[string]any); ok2 {
		mergePostedMarksFromIn(inMap, e.Name, items)
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		slog.Error("InputBulkSubmissionMarks marshal failed", "error", err, "key", e.Key)
		itemsJSON = []byte("[]")
	}

	label := e.Label
	if label == "" {
		label = "Marks by assignment"
	}

	nameLit, err := json.Marshal(e.Name)
	if err != nil {
		slog.Error("InputBulkSubmissionMarks marshal name failed", "error", err, "key", e.Key)
		nameLit = []byte(`""`)
	}
	initJS := fmt.Sprintf(`
$el.closest('form').addEventListener('submit', (ev) => {
	const d = Alpine.$data($el);
	if (!d || !Array.isArray(d.items)) return;
	const out = d.items.map(item => ({ id: Number(item.id) || 0, marks: Math.trunc(Number(item.marks) || 0) }));
	const h = $el.querySelector('input[type="hidden"][name=%s]');
	if (h) h.value = JSON.stringify(out);
}, true);
`, string(nameLit))

	if len(items) == 0 {
		return Div(Class("my-1 flex flex-col gap-2"),
			Div(Class("text-sm opacity-80"), Text("No assignment submissions to mark for this record.")),
			Attr("x-data", "{ items: [] }"),
			Input(Type("hidden"), Name(e.Name), Value("[]")),
		)
	}

	var rows []Node
	for i := range items {
		i := i
		title := items[i].AssignmentTitle
		if title == "" {
			title = fmt.Sprintf("Submission %d", items[i].ID)
		}
		subLine := title
		if items[i].CourseName != "" {
			subLine = title + " — " + items[i].CourseName
		}
		maxStr := "—"
		if items[i].MaxMarks > 0 {
			maxStr = fmt.Sprintf("%d", items[i].MaxMarks)
		}
		rows = append(rows, Div(Class("grid grid-cols-1 @md:grid-cols-2 gap-2 items-end border-b border-base-300 pb-2"),
			Div(Class("flex flex-col gap-0.5"),
				Div(Class("text-sm font-medium"), Text(subLine)),
				Div(Class("text-xs opacity-70"), Text(fmt.Sprintf("Max marks: %s", maxStr))),
			),
			Label(Class("form-control w-full max-w-xs"),
				Span(Class("label-text text-sm"), Text("Marks")),
				Input(
					Type("number"),
					Class("input input-bordered input-sm w-full"),
					Attr("x-model.number", fmt.Sprintf("items[%d].marks", i)),
					Attr("min", "0"),
					If(
						items[i].MaxMarks > 0,
						Attr("max", fmt.Sprint(items[i].MaxMarks)),
					),
				),
			),
		))
	}

	return Div(Class("my-1 flex flex-col gap-3"),
		Label(Class("label text-sm font-bold"), Text(label)),
		Div(
			Class("flex flex-col gap-3"),
			Attr("x-data", fmt.Sprintf(`{ items: %s }`, string(itemsJSON))),
			Attr("x-init", initJS),
			Group(rows),
			Input(Type("hidden"), Name(e.Name)),
		),
	)
}

func mergePostedMarksFromIn(inMap map[string]any, fieldName string, items []bulkMarksItem) {
	raw, has := inMap[fieldName]
	if !has {
		return
	}
	byID := make(map[uint]int)
	switch t := raw.(type) {
	case []bulkSubmissionMarksFormRow:
		for _, r := range t {
			byID[r.ID] = r.Marks
		}
	}
	if len(byID) == 0 {
		return
	}
	for i := range items {
		if m, ok := byID[items[i].ID]; ok {
			items[i].Marks = m
		}
	}
}

func (e InputBulkSubmissionMarks) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		return nil, nil
	}
	raw := strings.TrimSpace(vals[0])
	var rows []bulkSubmissionMarksFormRow
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, fmt.Errorf("invalid marks json: %w", err)
	}
	out := make([]bulkSubmissionMarksFormRow, 0, len(rows))
	seen := map[uint]struct{}{}
	for _, r := range rows {
		if r.ID == 0 {
			continue
		}
		if _, dup := seen[r.ID]; dup {
			continue
		}
		seen[r.ID] = struct{}{}
		out = append(out, r)
	}
	return out, nil
}

func (e InputBulkSubmissionMarks) GetName() string { return e.Name }
