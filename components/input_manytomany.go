package components

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputManyToMany[T any] struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter[[]uint]
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
	initialItems := []registry.Pair[uint, string]{}
	if e.Getter != nil {
		values, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputManyToMany getter failed", "error", err, "key", e.Key)
		} else {
			initialItems, err = e.getItems(values, ctx)
			if err != nil {
				slog.Error("InputManyToMany getter failed", "error", err, "key", e.Key)
			}
		}
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
	}

	itemsJson := "["

	for _, item := range initialItems {
		itemsJson += item.ToKVJson() + ","
	}

	itemsJson += "]"

	alpineData := fmt.Sprintf("{ items: %s, placeholder: '%s' }", itemsJson, placeholder)
	modalContainerId := fmt.Sprintf("fk-modal-%s", e.Name)
	eventHandler := fmt.Sprintf(
		"if ($event.detail.name === '%s') { "+
			"let idx = items.findIndex(i => Object.keys(i)[0] === String($event.detail.value)); "+
			"if (idx >= 0) { items.splice(idx, 1) } "+
			"else { items.push({[$event.detail.value]: $event.detail.display}) } "+
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
		Template(
			Attr("x-for", "(item, i) in items"),
			Attr(":key", "Object.keys(item)[0]"),
			Input(Type("hidden"), Attr(":name", fmt.Sprintf("'%s'", e.Name)), Attr(":value", "Object.keys(item)[0]")),
		),

		// Display area: clickable to open the selection modal
		Div(
			Class("flex flex-wrap gap-2 min-h-[2.5rem] p-2 rounded-lg border border-base-300 cursor-pointer"),
			Attr("hx-get", urlStr),
			Attr("hx-target", "#"+modalContainerId),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),

			// Show placeholder when empty
			Span(
				Attr("x-show", "items.length === 0"),
				Class("opacity-50"),
				Attr("x-text", "placeholder"),
			),

			// Render selected chips
			Template(
				Attr("x-for", "(item, i) in items"),
				Attr(":key", "Object.keys(item)[0]"),
				Div(
					Class("flex items-center gap-1 bg-base-200 rounded-lg px-3 py-1"),
					Attr("@click.stop", ""),
					Span(Class("text-sm"), Attr("x-text", "Object.values(item)[0]")),
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
		Div(Attr("id", modalContainerId), Class("fk-modal-container"),
			Attr("x-init", "document.body.appendChild($el)")),
	)
}

func (e InputManyToMany[T]) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	ids := make([]uint, 0, len(vals))
	idErrors := make([]error, 0, len(vals))
	for _, s := range vals {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		modelValue := new(T)
		db := ctx.Value("$db").(*gorm.DB)
		if err := db.Model(modelValue).Where("ID = ?", i).First(modelValue).Error; err != nil {
			slog.Error("Error while fetching data for the specified foreign key", "error", err)
			idErrors = append(idErrors, err)
		}
		ids = append(ids, uint(i))
	}
	if len(idErrors) > 0 {
		return nil, errors.Join(idErrors...)
	}
	return ids, nil
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

func (e InputManyToMany[T]) getItems(ids []uint, ctx context.Context) ([]registry.Pair[uint, string], error) {
	pairs := make([]registry.Pair[uint, string], 0, len(ids))
	if len(ids) == 0 {
		return pairs, nil
	}

	dbVal := ctx.Value("$db")
	if dbVal == nil {
		return nil, errors.New("missing $db in context")
	}
	db, ok := dbVal.(*gorm.DB)
	if !ok {
		return nil, errors.New("context $db is not a *gorm.DB")
	}

	type row struct {
		ID    uint
		Value string
	}
	rows := []row{}

	displayAttr := e.DisplayAttr
	if displayAttr == "" {
		displayAttr = "id"
	}

	// Select only ID and DisplayAttr
	query := db.Model(new(T)).
		Select("id, "+displayAttr+" as value").
		Where("id IN ?", ids).
		Scan(&rows)
	if query.Error != nil {
		return nil, query.Error
	}

	foundIDs := make(map[uint]bool, len(rows))
	for _, r := range rows {
		pairs = append(pairs, registry.Pair[uint, string]{Key: r.ID, Value: r.Value})
		foundIDs[r.ID] = true
	}

	if len(pairs) < len(ids) {
		missing := make([]uint, 0, len(ids)-len(pairs))
		for _, id := range ids {
			if !foundIDs[id] {
				missing = append(missing, id)
			}
		}
		return pairs, fmt.Errorf("the following IDs were not found: %v", missing)
	}

	return pairs, nil
}
