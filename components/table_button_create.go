package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
)

const (
	tableButtonCreateDefaultIcon    = "plus"
	tableButtonCreateDefaultClasses = "btn-square btn-outline btn-sm"
)

// TableButtonCreate is the default “add row” control for DataTable.Actions.
// Icon and Classes match the former DataTable+ButtonLink defaults; set Icon or Classes non-empty to override.
type TableButtonCreate struct {
	Page
	Link        getters.Getter[string]
	Label       string
	GetterLabel getters.Getter[string]
	Icon        string
	IconClasses string
	Classes     string
}

func (e TableButtonCreate) GetKey() string {
	return e.Key
}

func (e TableButtonCreate) GetRoles() []string {
	return e.Roles
}

func (e TableButtonCreate) Build(ctx context.Context) gomponents.Node {
	icon := e.Icon
	if icon == "" {
		icon = tableButtonCreateDefaultIcon
	}
	classes := e.Classes
	if classes == "" {
		classes = tableButtonCreateDefaultClasses
	}
	return ButtonLink{
		Page:        e.Page,
		Label:       e.Label,
		GetterLabel: e.GetterLabel,
		Link:        e.Link,
		Icon:        icon,
		IconClasses: e.IconClasses,
		Classes:     classes,
	}.Build(ctx)
}
