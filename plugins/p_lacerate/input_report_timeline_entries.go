package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	g "maragu.dev/gomponents/html"
)

var _ components.InputInterface = InputTimelineEntries{}

// InputTimelineEntries renders repeatable timeline entry fields and posts one JSON payload.
type InputTimelineEntries struct {
	components.Page
	Label   string
	Name    string
	Getter  getters.Getter[string]
	Classes string
}

func (e InputTimelineEntries) GetKey() string {
	return e.Key
}

func (e InputTimelineEntries) GetRoles() []string {
	return e.Roles
}

func (e InputTimelineEntries) Build(ctx context.Context) Node {
	items := []reportTimelineEntryFormInput{{}}
	if e.Getter != nil {
		raw, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTimelineEntries getter failed", "error", err, "key", e.Key)
		} else if strings.TrimSpace(raw) != "" {
			var parsed []reportTimelineEntryFormInput
			if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
				slog.Error("InputTimelineEntries unmarshal failed", "error", err, "key", e.Key)
			} else if len(parsed) != 0 {
				items = parsed
			}
		}
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		slog.Error("InputTimelineEntries marshal failed", "error", err, "key", e.Key)
		itemsJSON = []byte(`[{"datetime":"","title":"","content":""}]`)
	}

	alpineData := fmt.Sprintf(`{
		items: %s,
		add() { this.items.push({datetime: '', title: '', content: ''}); },
		remove(i) {
			this.items.splice(i, 1);
			if (this.items.length === 0) this.add();
		}
	}`, itemsJSON)
	initJS := fmt.Sprintf(`
$el.closest('form').addEventListener('submit', (e) => {
	const d = Alpine.$data($el);
	if (!d || !Array.isArray(d.items)) return;
	const cleaned = d.items.map(item => ({
		datetime: String(item.datetime || '').trim(),
		title: String(item.title || '').trim(),
		content: String(item.content || '').trim(),
	})).filter(item => item.datetime !== '' || item.title !== '' || item.content !== '');
	const h = $el.querySelector('input[type="hidden"][name=%s]');
	if (h) h.value = JSON.stringify(cleaned);
}, true);
`, strconv.Quote(e.Name))

	return g.Div(g.Class(fmt.Sprintf("my-1 %s", e.Classes)),
		g.Label(g.Class("label text-sm font-bold"), Text(e.Label)),
		g.Div(
			g.Class("flex flex-col gap-3"),
			Attr("x-data", alpineData),
			Attr("x-init", initJS),
			g.Template(
				Attr("x-for", "(item, i) in items"),
				Attr(":key", "i"),
				g.Div(
					g.Class("border border-base-300 rounded p-4 flex flex-col gap-3"),
					g.Div(
						g.Class("flex items-center justify-between gap-2"),
						g.Div(g.Class("font-medium text-sm"), Text("Timeline entry")),
						g.Button(
							g.Type("button"),
							g.Class("btn btn-ghost btn-sm"),
							Attr("@click", "remove(i)"),
							Text("Remove"),
						),
					),
					g.Div(
						g.Class("grid gap-3 md:grid-cols-2"),
						g.Label(
							g.Class("label text-sm font-bold flex flex-col items-start gap-1"),
							Text("Datetime"),
							g.Input(
								g.Type("datetime-local"),
								g.Class("input input-bordered w-full"),
								Attr("x-model", "items[i].datetime"),
							),
						),
						g.Label(
							g.Class("label text-sm font-bold flex flex-col items-start gap-1"),
							Text("Title"),
							g.Input(
								g.Type("text"),
								g.Class("input input-bordered w-full"),
								Attr("x-model", "items[i].title"),
							),
						),
					),
					g.Label(
						g.Class("label text-sm font-bold flex flex-col items-start gap-1"),
						Text("Content"),
						g.Textarea(
							g.Class("textarea textarea-bordered w-full min-h-32"),
							g.Rows("6"),
							Attr("x-model", "items[i].content"),
						),
					),
				),
			),
			g.Button(
				g.Type("button"),
				g.Class("btn btn-outline btn-sm self-start"),
				Attr("@click", "add()"),
				Text("Add entry"),
			),
			g.Input(g.Type("hidden"), g.Name(e.Name)),
		),
	)
}

func (e InputTimelineEntries) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

func (e InputTimelineEntries) GetName() string {
	return e.Name
}
