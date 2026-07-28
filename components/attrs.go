package components

import (
	"context"
	"html"
	"html/template"
	"strings"

	"github.com/lariv-in/lariv/getters"
)

// HTMLAttributes is a map of HTML attribute names to values used by templates and Attr getters.
// It is an alias of map[string]string so getters returning Getter[map[string]string] assign cleanly.
type HTMLAttributes = map[string]string

// MergeAttrs returns a new map with values from extras overwriting base.
func MergeAttrs(base HTMLAttributes, extras ...HTMLAttributes) HTMLAttributes {
	out := HTMLAttributes{}
	for k, v := range base {
		out[k] = v
	}
	for _, m := range extras {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// ResolveAttrs evaluates an optional Attr getter, returning an empty map when nil.
func ResolveAttrs(ctx context.Context, g getters.Getter[HTMLAttributes]) (HTMLAttributes, error) {
	if g == nil {
		return HTMLAttributes{}, nil
	}
	attrs, err := g(ctx)
	if err != nil {
		return nil, err
	}
	if attrs == nil {
		return HTMLAttributes{}, nil
	}
	return attrs, nil
}

func renderAttrs(m HTMLAttributes) template.HTMLAttr {
	if len(m) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range m {
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(v))
		b.WriteByte('"')
	}
	return template.HTMLAttr(b.String())
}
