package components

import (
	"context"
	"reflect"
	"slices"

	"maragu.dev/gomponents"
)

type PageInterface interface {
	Build(context.Context) gomponents.Node
	GetKey() string
	GetRoles() []string
}

// Page struct defines fields that are common in all components
type Page struct {
	Key   string
	Roles []string
}

func Render(p PageInterface, ctx context.Context) gomponents.Node {
	roles := GetRequiredRoles(p)
	currentRole, _ := ctx.Value("$role").(string)
	if roles == nil {
		return p.Build(ctx)
	}

	if slices.Contains(roles, currentRole) {
		return p.Build(ctx)
	}
	return gomponents.Group{}
}

func GetRequiredRoles(p PageInterface) []string {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	page, ok := v.FieldByName("Page").Interface().(Page)
	if !ok {
		return nil
	}
	return page.Roles
}
