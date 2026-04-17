package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputForeignKey[T any] struct {
	Page
	Label       string
	Name        string
	Getter      getters.Getter[T]
	Display     getters.Getter[string]
	Placeholder string
	Url         getters.Getter[string]
	Required    bool
	Classes     string
	Attr        getters.Getter[Node]
	// Hidden renders only a hidden input (no label or picker). Use for carried IDs.
	Hidden bool
}

func (e InputForeignKey[T]) GetKey() string {
	return e.Key
}

func (e InputForeignKey[T]) GetRoles() []string {
	return e.Roles
}

func (e InputForeignKey[T]) Build(ctx context.Context) Node {
	valuePk := ""
	displayValue := ""

	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputForeignKey getter failed", "error", err, "key", e.Key)
		} else {
			valueMap := getters.MapFromStruct(value)
			if len(valueMap) > 0 {
				haveSelectedID := false
				if idVal, exists := valueMap["ID"]; exists {
					if rv := reflect.ValueOf(idVal); rv.IsValid() && !rv.IsZero() {
						valuePk = fmt.Sprintf("%v", idVal)
						haveSelectedID = true
					}
				} else if idVal, exists := valueMap["id"]; exists {
					if rv := reflect.ValueOf(idVal); rv.IsValid() && !rv.IsZero() {
						valuePk = fmt.Sprintf("%v", idVal)
						haveSelectedID = true
					}
				}
				if e.Display != nil && haveSelectedID {
					displayStr, err := e.Display(context.WithValue(ctx, "$in", valueMap))
					if err != nil {
						slog.Error("InputForeignKey display getter failed", "error", err, "key", e.Key)
					} else {
						displayValue = displayStr
					}
				}
			}
		}
	}

	if e.Hidden {
		wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
		wrapClass += " hidden"
		return Div(Class(wrapClass),
			Input(Type("hidden"), Name(e.Name), Value(valuePk),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputForeignKey attr getter failed", "error", err, "key", e.Key)
						return out
					}
					return n
				}),
			),
		)
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
			slog.Error("InputForeignKey url getter failed", "error", err, "key", e.Key)
			urlStr = ""
		}
	}

	alpinePayload, errAlpine := json.Marshal(map[string]string{
		"value":       valuePk,
		"display":     displayValue,
		"placeholder": placeholder,
	})
	if errAlpine != nil {
		alpinePayload = []byte(`{"value":"","display":"","placeholder":""}`)
	}
	alpineData := string(alpinePayload)
	// Selector dialog is closed from the table row @click (getters.Select); avoid removing the wrong dialog here.
	eventHandler := fmt.Sprintf("if ($event.detail.name === '%s') { value = $event.detail.value; display = $event.detail.display }", e.Name)

	return Div(
		Class(fmt.Sprintf("my-1 relative %s", e.Classes)),
		Attr("x-data", alpineData),
		Attr("@fk-select.window", eventHandler),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("hidden"), Name(e.Name), Attr(":value", "value"),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					defer func() {
						if r := recover(); r != nil {
							slog.Error("InputForeignKey attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputForeignKey attr getter failed", "error", err, "key", e.Key)
						return out
					}
					if n == nil {
						return out
					}
					v := reflect.ValueOf(n)
					if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Interface || v.Kind() == reflect.Func) && v.IsNil() {
						return out
					}
					return n
				}),
			),
			Div(Class("flex w-full items-stretch gap-1"),
				Div(Class("input input-bordered flex-1 flex items-center cursor-pointer"),
					Attr(":class", "display ? '' : 'opacity-50'"),
					Attr("hx-get", urlStr),
					Attr("hx-target", HTMXTargetBodyModal),
					Attr("hx-swap", HTMXSwapBodyModal),
					Attr("hx-push-url", "false"),
					El("span", Attr("x-text", "display || placeholder")),
				),
				If(!e.Required,
					Button(
						Type("button"),
						Class("btn btn-ghost btn-square shrink-0"),
						Attr("@click.stop", "value = ''; display = ''"),
						Attr("x-show", "value"),
						Attr("aria-label", "Clear selection"),
						Render(Icon{Name: "x-mark"}, ctx),
					),
				),
			),
		),
	)
}

func (e InputForeignKey[T]) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	if strings.TrimSpace(vals[0]) == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, err
	}
	modelValue := new(T)

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("InputForeignKey: db from context", "error", err)
		return nil, err
	}

	row, err := gorm.G[T](db).Where("ID = ?", i).First(ctx)
	if err != nil {
		slog.Error("Error while fetching data for the specified foreign key", "error", err)
		return nil, err
	}
	*modelValue = row

	return uint(i), nil
}

func (e InputForeignKey[T]) GetName() string {
	return e.Name
}
