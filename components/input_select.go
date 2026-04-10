package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputSelect renders a <select> bound to a form field. Choices lists option
// keys and labels; Getter supplies the current selection as a Pair (Key is the
// stored value, Value is the label). Parse returns the submitted option Key as T.
// When Required is false, a first option with value "" is shown (label from EmptyLabel or "—").
// T must be comparable because it is the Key type of registry.Pair.
type InputSelect[T comparable] struct {
	Page
	Label    string
	Name     string
	Choices  getters.Getter[[]registry.Pair[T, string]]
	Getter   getters.Getter[registry.Pair[T, string]]
	Required bool
	// EmptyLabel is the visible label for the empty value option when Required is false.
	// If empty, "—" is used.
	EmptyLabel string
	Classes    string
	Hidden     bool
	Attr       getters.Getter[Node]
}

func (e InputSelect[T]) GetKey() string {
	return e.Key
}

func (e InputSelect[T]) GetRoles() []string {
	return e.Roles
}

func (e InputSelect[T]) Build(ctx context.Context) Node {
	var zero T

	choices := []registry.Pair[T, string]{}
	if e.Choices != nil {
		opts, err := e.Choices(ctx)
		if err != nil {
			slog.Error("InputSelect Choices getter failed", "error", err, "key", e.Key)
		} else {
			choices = opts
		}
	}

	rawSel := ""
	if e.Getter != nil {
		pair, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputSelect Getter failed", "error", err, "key", e.Key)
		} else if any(pair.Key) != any(zero) {
			rawSel = fmt.Sprint(pair.Key)
		}
	}

	emptyLab := "—"
	if e.EmptyLabel != "" {
		emptyLab = e.EmptyLabel
	}

	optionNodes := []Node{}
	if !e.Required {
		optionNodes = append(optionNodes,
			Option(Value(""), If(rawSel == "", Attr("selected", "")), Text(emptyLab)),
		)
	}
	for _, opt := range choices {
		ks := fmt.Sprint(opt.Key)
		optionNodes = append(optionNodes,
			Option(Value(ks), If(rawSel == ks, Attr("selected", "")), Text(opt.Value)),
		)
	}

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	return Div(Class(wrapClass),
		Label(Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Select(
				Name(e.Name),
				Class(fmt.Sprintf("select select-bordered w-full %s", e.Classes)),
				Group(optionNodes),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() Node {
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputSelect Attr getter failed", "error", err, "key", e.Key)
						return Raw("")
					}
					return n
				}),
			),
		),
	)
}

func (e InputSelect[T]) Parse(v any, ctx context.Context) (any, error) {
	var zero T

	vals, ok := v.([]string)
	if !ok || len(vals) == 0 || vals[0] == "" {
		return zero, nil
	}
	submitted := vals[0]

	if e.Choices == nil {
		return zero, fmt.Errorf("InputSelect: no Choices getter for validation")
	}
	choices, err := e.Choices(ctx)
	if err != nil {
		return zero, fmt.Errorf("InputSelect: Choices getter failed: %w", err)
	}
	for _, opt := range choices {
		if fmt.Sprint(opt.Key) == submitted {
			return opt.Key, nil
		}
	}
	return zero, fmt.Errorf("InputSelect: invalid choice %q", submitted)
}

func (e InputSelect[T]) GetName() string {
	return e.Name
}
