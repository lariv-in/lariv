package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputManyToMany[T any] struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter[T]
	DisplayAttr string
	Placeholder string
	Url         getters.Getter[string]
	Required    bool
	Classes     string
}

func (e InputManyToMany[T]) GetKey() string {
	return e.Key
}

func (e InputManyToMany[T]) GetRoles() []string {
	return e.Roles
}

func (e InputManyToMany[T]) Build(ctx context.Context) Node {
	initialItems := []string{}
	if e.Getter != nil {
		values, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputManyToMany getter failed", "error", err, "key", e.Key)
		} else {
			for _, item := range e.toItemMaps(values) {
				pk, ok := itemID(item)
				if !ok {
					continue
				}
				display := e.itemDisplay(item)
				initialItems = append(initialItems, fmt.Sprintf("{value:%s,display:%s}", jsQuote(pk), jsQuote(display)))
			}
		}
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
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

	urlStr := ""
	if e.Url != nil {
		var err error
		urlStr, err = e.Url(ctx)
		if err != nil {
			slog.Error("InputManyToMany url getter failed", "error", err, "key", e.Key)
			urlStr = ""
		}
	}

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
			Attr("hx-get", urlStr),
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

func (e InputManyToMany[T]) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	// Return the full slice of selected IDs
	return vals, nil
}

func (e InputManyToMany[T]) GetName() string {
	return e.Name
}

func (e InputManyToMany[T]) toItemMaps(values T) []map[string]any {
	rv := reflect.ValueOf(values)
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil
	}

	items := make([]map[string]any, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := getters.MapFromStruct(rv.Index(i).Interface())
		if len(item) == 0 {
			continue
		}
		items = append(items, item)
	}
	return items
}

func itemID(item map[string]any) (string, bool) {
	if id, ok := item["ID"]; ok && id != nil {
		return fmt.Sprintf("%v", id), true
	}
	if id, ok := item["id"]; ok && id != nil {
		return fmt.Sprintf("%v", id), true
	}
	return "", false
}

func (e InputManyToMany[T]) itemDisplay(item map[string]any) string {
	if e.DisplayAttr == "" {
		return ""
	}
	if d, ok := item[e.DisplayAttr]; ok && d != nil {
		return fmt.Sprintf("%v", d)
	}
	return ""
}

func jsQuote(v string) string {
	return strconv.Quote(v)
}
