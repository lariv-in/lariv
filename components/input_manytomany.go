package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputManyToMany struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter
	DisplayAttr string
	Placeholder string
	Url         getters.Getter
	Required    bool
	Classes     string
}

func (e InputManyToMany) GetKey() string {
	return e.Key
}

func (e InputManyToMany) GetRoles() []string {
	return e.Roles
}

func (e InputManyToMany) Build(ctx context.Context) Node {
	values := getters.IfOrGetter(e.Getter, ctx, nil)

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
	}

	// Build initial items JSON array from existing values
	// values could be a slice of maps (from GetterAssociation or similar)
	var initialItems []string
	if values != nil {
		switch v := values.(type) {
		case []map[string]any:
			for _, item := range v {
				pk := fmt.Sprintf("%v", item["ID"])
				if pk == "<nil>" {
					pk = fmt.Sprintf("%v", item["id"])
				}
				display := ""
				if e.DisplayAttr != "" {
					if d, ok := item[e.DisplayAttr]; ok {
						display = fmt.Sprintf("%v", d)
					}
				}
				initialItems = append(initialItems, fmt.Sprintf("{value:'%s',display:'%s'}", pk, display))
			}
		case []any:
			for _, raw := range v {
				if item, ok := raw.(map[string]any); ok {
					pk := fmt.Sprintf("%v", item["ID"])
					if pk == "<nil>" {
						pk = fmt.Sprintf("%v", item["id"])
					}
					display := ""
					if e.DisplayAttr != "" {
						if d, ok := item[e.DisplayAttr]; ok {
							display = fmt.Sprintf("%v", d)
						}
					}
					initialItems = append(initialItems, fmt.Sprintf("{value:'%s',display:'%s'}", pk, display))
				}
			}
		}
	}

	itemsJson := "[" + strings.Join(initialItems, ",") + "]"

	alpineData := fmt.Sprintf("{ items: %s, placeholder: '%s' }", itemsJson, placeholder)
	eventHandler := fmt.Sprintf(
		"if ($event.detail.name === '%s') { "+
			"if (!items.some(i => i.value === String($event.detail.value))) { "+
			"items.push({value: String($event.detail.value), display: String($event.detail.display)}); "+
			"} "+
			"$el.querySelector('.fk-modal-container').innerHTML = ''; "+
			"}",
		e.Name,
	)

	removeHandler := "items = items.filter((_, j) => j !== i)"

	url := fmt.Sprintf("%v", getters.IfOrGetter(e.Url, ctx, ""))

	return Div(
		Class(fmt.Sprintf("my-1 relative %s", e.Classes)),
		Attr("x-data", alpineData),
		Attr("@fk-multi-select.window", eventHandler),
		Label(Class("label text-sm font-bold"), Text(e.Label)),

		// Hidden inputs: one per selected item, rendered by Alpine template
		El("template",
			Attr("x-for", "(item, i) in items"),
			Attr(":key", "item.value"),
			Input(Type("hidden"), Attr(":name", fmt.Sprintf("'%s'", e.Name)), Attr(":value", "item.value")),
		),

		// Display area: clickable to open the selection modal
		Div(
			Class("flex flex-wrap gap-2 min-h-[2.5rem] p-2 rounded-lg border border-base-300 cursor-pointer"),
			Attr("hx-get", url),
			Attr("hx-target", "next .fk-modal-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),

			// Show placeholder when empty
			El("span",
				Attr("x-show", "items.length === 0"),
				Class("opacity-50"),
				Attr("x-text", "placeholder"),
			),

			// Render selected chips
			El("template",
				Attr("x-for", "(item, i) in items"),
				Attr(":key", "item.value"),
				Div(
					Class("flex items-center gap-1 bg-base-200 rounded-lg px-3 py-1"),
					Attr("@click.stop", ""),
					El("span", Class("text-sm"), Attr("x-text", "item.display")),
					Button(
						Type("button"),
						Class("btn btn-ghost btn-xs"),
						Attr("@click.stop", removeHandler),
						Text("✕"),
					),
				),
			),
		),

		// Modal container for the selection table
		Div(Class("fk-modal-container")),
	)
}

func (e InputManyToMany) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	// Return the full slice of selected IDs
	return vals, nil
}

func (e InputManyToMany) GetName() string {
	return e.Name
}
