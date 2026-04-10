package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputStringList edits a JSON array of strings (posted as one hidden field).
// It follows the same submit pattern as InputKeyValue: Alpine state + JSON on form submit.
type InputStringList struct {
	Page
	Label   string
	Name    string
	Getter  getters.Getter[datatypes.JSON]
	Classes string
}

func (e InputStringList) GetKey() string {
	return e.Key
}

func (e InputStringList) GetRoles() []string {
	return e.Roles
}

func (e InputStringList) Build(ctx context.Context) Node {
	var items []string
	if e.Getter != nil {
		j, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputStringList Getter failed", "error", err, "key", e.Key)
		} else if len(j) > 0 {
			if err := json.Unmarshal(j, &items); err != nil {
				slog.Error("InputStringList unmarshal failed", "error", err, "key", e.Key)
			}
		}
	}
	if len(items) == 0 {
		items = []string{""}
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		itemsJSON = []byte(`[""]`)
	}

	alpineData := fmt.Sprintf(`{
		items: %s,
		add() { this.items.push(''); },
		remove(i) { this.items.splice(i, 1); if (this.items.length === 0) this.items.push(''); }
	}`, itemsJSON)

	// Capture phase so this runs before Alpine @submit.prevent on the form dispatches
	// "lago-form-submit" (e.g. FormListenBoostedPost), which reads the form via htmx.values
	// before bubble-phase submit handlers run.
	initJS := fmt.Sprintf(`
$el.closest('form').addEventListener('submit', (e) => {
	const d = Alpine.$data($el);
	if (!d || !Array.isArray(d.items)) return;
	const cleaned = d.items.map(s => String(s).trim()).filter(s => s !== '');
	const h = $el.querySelector('input[type="hidden"][name=%s]');
	if (h) h.value = JSON.stringify(cleaned);
}, true);
`, strconv.Quote(e.Name))

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	return Div(Class(wrapClass),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Div(
				Attr("x-data", alpineData),
				Attr("x-init", initJS),
				Template(
					Attr("x-for", "(item, i) in items"),
					Attr(":key", "i"),
					Div(
						Class("flex gap-2 items-center my-1"),
						Input(
							Type("text"),
							Class("input input-bordered flex-1"),
							Attr("x-model", "items[i]"),
							Attr("placeholder", "Option value"),
						),
						Button(
							Type("button"),
							Class("btn btn-ghost btn-sm shrink-0"),
							Attr("@click", "remove(i)"),
							Text("Remove"),
						),
					),
				),
				Button(
					Type("button"),
					Class("btn btn-outline btn-sm mt-1"),
					Attr("@click", "add()"),
					Text("Add option"),
				),
				Input(Type("hidden"), Name(e.Name)),
			),
		),
	)
}

func (e InputStringList) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		return "[]", nil
	}
	raw := strings.TrimSpace(vals[0])
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil, fmt.Errorf("options must be a JSON array of strings: %w", err)
	}
	return raw, nil
}

func (e InputStringList) GetName() string {
	return e.Name
}
