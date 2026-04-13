package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AssociationIDs marks a parsed many-to-many form value so CRUD handlers can
// persist it through GORM association APIs instead of treating it as a column.
type AssociationIDs struct {
	Field string
	IDs   []uint
}

type InputManyToMany[T any] struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter[[]T]
	Display     getters.Getter[string]
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
	items := e.initialSelections(ctx)
	if items == nil {
		items = []registry.Pair[string, string]{}
	}

	placeholder := e.Placeholder
	if placeholder == "" {
		placeholder = "Select..."
	}

	urlStr := ""
	if e.Url != nil {
		var err error
		urlStr, err = e.Url(ctx)
		if err != nil {
			slog.Error("InputManyToMany url getter failed", "error", err, "key", e.Key)
		}
	}
	if urlStr != "" {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			query := parsedURL.Query()
			query.Set("target_input", e.Name)
			parsedURL.RawQuery = query.Encode()
			urlStr = parsedURL.String()
		}
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		slog.Error("InputManyToMany items marshal failed", "error", err, "key", e.Key)
		itemsJSON = []byte("[]")
	}
	nameJSON, _ := json.Marshal(e.Name)
	placeholderJSON, _ := json.Marshal(placeholder)

	alpineData := fmt.Sprintf(`{
		items: %s,
		placeholder: %s,
		syncStore() {
			if (typeof Alpine === 'undefined') {
				return
			}
			if (!Alpine.store('m2mSelections')) {
				Alpine.store('m2mSelections', {})
			}
			Alpine.store('m2mSelections')[%s] = this.items
		},
		hasItem(value) {
			value = String(value)
			return this.items.some(item => item.Key === value)
		},
		addItem(detail) {
			const value = String(detail.value)
			if (this.hasItem(value)) {
				return
			}
			const display = detail.display ? String(detail.display) : value
			this.items = [...this.items, { Key: value, Value: display }]
			this.syncStore()
		},
		removeItem(detail) {
			const value = String(detail.value)
			this.items = this.items.filter(item => item.Key !== value)
			this.syncStore()
		},
		eventHandler(ev) {
			if (ev.detail.name === %s) {
				if (!this.hasItem(ev.detail.value)) {
					this.addItem(ev.detail)
				} else {
					this.removeItem(ev.detail)
				}
			}
		}
	}`, itemsJSON, placeholderJSON, string(nameJSON), string(nameJSON))
	eventHandler := "eventHandler($event)"

	return Div(
		Class(fmt.Sprintf("my-1 relative %s", e.Classes)),
		Attr("x-data", alpineData),
		Attr("x-init", "syncStore()"),
		Attr("@fk-multi-select.window", eventHandler),
		Div(Class("flex flex-col items-start gap-1"),
			If(e.Label != "", Label(Class("label text-sm font-bold"), Text(e.Label))),
			Div(
				Class("input input-bordered w-full min-h-12 h-auto flex flex-wrap items-center gap-2 cursor-pointer"),
				Attr(":class", "items.length ? '' : 'opacity-50'"),
				Attr("hx-get", urlStr),
				Attr("hx-target", HTMXTargetBodyModal),
				Attr("hx-swap", HTMXSwapBodyModal),
				Attr("hx-push-url", "false"),
				Span(
					Attr("x-show", "items.length === 0"),
					Attr("x-text", "placeholder"),
				),
				Template(
					Attr("x-for", "item in items"),
					Attr(":key", "item.Key"),
					Div(
						Class("flex items-center gap-1 rounded-lg bg-base-200 pl-2 pr-1 py-1"),
						Attr("@click", "$event.stopPropagation()"),
						Input(Type("hidden"), Name(e.Name), Attr(":value", "item.Key")),
						Span(Class("text-sm flex-1 min-w-0 truncate"), Attr("x-text", "item.Value")),
						Button(
							Type("button"),
							Class("btn btn-ghost btn-square btn-xs shrink-0"),
							Attr("@click.stop", "removeItem({ value: item.Key })"),
							Attr("aria-label", "Remove"),
							Render(Icon{Name: "x-mark"}, ctx),
						),
					),
				),
			),
		),
	)
}

func (e InputManyToMany[T]) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	ids := make([]uint, 0, len(vals))
	seen := map[uint]struct{}{}
	for _, raw := range vals {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return AssociationIDs{Field: e.Name, IDs: ids}, err
		}
		if _, exists := seen[uint(id)]; exists {
			continue
		}
		seen[uint(id)] = struct{}{}
		ids = append(ids, uint(id))
	}

	if e.Required && len(ids) == 0 {
		return AssociationIDs{Field: e.Name, IDs: ids}, fmt.Errorf("Please select at least one value")
	}

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return AssociationIDs{Field: e.Name, IDs: ids}, err
	}

	if len(ids) > 0 {
		count, err := gorm.G[T](db).Where("id IN ?", ids).Count(ctx, "*")
		if err != nil {
			slog.Error("Error while fetching data for the specified many-to-many values", "error", err)
			return AssociationIDs{Field: e.Name, IDs: ids}, err
		}
		if count != int64(len(ids)) {
			return AssociationIDs{Field: e.Name, IDs: ids}, fmt.Errorf("One or more selected values are invalid")
		}
	}

	return AssociationIDs{Field: e.Name, IDs: ids}, nil
}

func (e InputManyToMany[T]) GetName() string {
	return e.Name
}

func (e InputManyToMany[T]) initialSelections(ctx context.Context) []registry.Pair[string, string] {
	if items, ok := e.selectionsFromContext(ctx); ok {
		return items
	}

	if e.Getter == nil {
		return nil
	}

	values, err := e.Getter(ctx)
	if err != nil {
		slog.Error("InputManyToMany getter failed", "error", err, "key", e.Key)
		return nil
	}

	items := make([]registry.Pair[string, string], 0, len(values))
	for _, value := range values {
		item, ok := e.selectionForValue(ctx, value)
		if ok {
			items = append(items, item)
		}
	}
	return items
}

func (e InputManyToMany[T]) selectionsFromContext(ctx context.Context) ([]registry.Pair[string, string], bool) {
	inMap, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
	if !ok {
		return nil, false
	}
	raw, ok := inMap[e.Name]
	if !ok {
		return nil, false
	}

	switch value := raw.(type) {
	case AssociationIDs:
		return e.selectionsForIDs(ctx, value.IDs), true
	case *AssociationIDs:
		if value == nil {
			return nil, true
		}
		return e.selectionsForIDs(ctx, value.IDs), true
	case []uint:
		return e.selectionsForIDs(ctx, value), true
	case []int:
		ids := make([]uint, 0, len(value))
		for _, id := range value {
			if id > 0 {
				ids = append(ids, uint(id))
			}
		}
		return e.selectionsForIDs(ctx, ids), true
	case []string:
		ids := make([]uint, 0, len(value))
		for _, rawID := range value {
			if rawID == "" {
				continue
			}
			id, err := strconv.ParseUint(rawID, 10, 64)
			if err != nil {
				continue
			}
			ids = append(ids, uint(id))
		}
		return e.selectionsForIDs(ctx, ids), true
	default:
		return nil, false
	}
}

func (e InputManyToMany[T]) selectionsForIDs(ctx context.Context, ids []uint) []registry.Pair[string, string] {
	if len(ids) == 0 {
		return nil
	}

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		items := make([]registry.Pair[string, string], 0, len(ids))
		for _, id := range ids {
			items = append(items, registry.Pair[string, string]{Key: strconv.FormatUint(uint64(id), 10), Value: strconv.FormatUint(uint64(id), 10)})
		}
		return items
	}

	values, err := gorm.G[T](db).Where("id IN ?", ids).Find(ctx)
	if err != nil {
		slog.Error("InputManyToMany lookup failed", "error", err, "key", e.Key)
		return nil
	}

	byID := map[string]registry.Pair[string, string]{}
	for _, value := range values {
		item, ok := e.selectionForValue(ctx, value)
		if ok {
			byID[item.Key] = item
		}
	}

	items := make([]registry.Pair[string, string], 0, len(ids))
	for _, id := range ids {
		key := strconv.FormatUint(uint64(id), 10)
		if item, ok := byID[key]; ok {
			items = append(items, item)
			continue
		}
		items = append(items, registry.Pair[string, string]{Key: key, Value: key})
	}
	return items
}

func (e InputManyToMany[T]) selectionForValue(ctx context.Context, value T) (registry.Pair[string, string], bool) {
	return manyToManySelectionPair(ctx, value, e.Display, e.Key)
}

// manyToManySelectionPair maps a related model value to id/display strings; shared by
// InputManyToMany and FieldManyToMany so detail and form views stay consistent.
func manyToManySelectionPair[T any](ctx context.Context, value T, display getters.Getter[string], logKey string) (registry.Pair[string, string], bool) {
	valueMap := getters.MapFromStruct(value)
	if len(valueMap) == 0 {
		return registry.Pair[string, string]{}, false
	}

	var rawID any
	var ok bool
	if rawID, ok = valueMap["ID"]; !ok {
		rawID, ok = valueMap["id"]
	}
	if !ok {
		return registry.Pair[string, string]{}, false
	}

	selection := registry.Pair[string, string]{Key: fmt.Sprintf("%v", rawID), Value: fmt.Sprintf("%v", rawID)}
	if display != nil {
		d, err := display(context.WithValue(ctx, getters.ContextKeyIn, valueMap))
		if err != nil {
			slog.Error("many-to-many display getter failed", "error", err, "key", logKey)
		} else if d != "" {
			selection.Value = d
		}
		return selection, true
	}

	if name, ok := valueMap["Name"]; ok {
		selection.Value = fmt.Sprintf("%v", name)
	}
	return selection, true
}
