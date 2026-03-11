package components

import (
	"context"
	"reflect"
	"slices"

	"maragu.dev/gomponents"
)

type PageInterface interface {
	Build(context.Context) gomponents.Node
}

type Page struct {
	RenderKeys []string
}

func Render(p PageInterface, ctx context.Context) gomponents.Node {
	keys := GetRenderKeys(p)

	if keys == nil {
		return p.Build(ctx)
	}

	if slices.Contains(keys, ctx.Value("$render_key").(string)) {
		return p.Build(ctx)
	}
	return gomponents.Group{}
}

func GetRenderKeys(p PageInterface) []string {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	page, ok := v.FieldByName("Page").Interface().(Page)
	if !ok {
		return nil
	}
	return page.RenderKeys
}
