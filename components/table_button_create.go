package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
)

// Target constants defining fallback defaults for creation controls.
const (
	// tableButtonCreateDefaultIcon is the default SVG icon identifier ("plus").
	tableButtonCreateDefaultIcon = "plus"
	// tableButtonCreateDefaultClasses is the default DaisyUI styling applied ("btn-square btn-outline btn-sm").
	tableButtonCreateDefaultClasses = "btn-square btn-outline btn-sm"
)

// TableButtonCreate represents the default "add record" trigger option for DataTable.Actions lists.
// It wraps a standard [ButtonLink] component to render a creation link featuring an optional add icon and tooltip label.
//
// Use Cases:
//   - Adding "Create New Record", "Add User", or "Import File" option buttons inside table toolbars.
//
// Example:
//
//	&components.TableButtonCreate{
//	    Link:  lago.RoutePath("admin.UserCreate", nil),
//	    Label: "Create Account",
//	}
type TableButtonCreate struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Link represents the dynamic function retrieving the destination anchor target path.
	Link getters.Getter[string]
	// Label represents the static string label text displayed on hover or inline.
	Label string
	// GetterLabel represents the dynamic function retrieving the display label text (takes precedence over Label).
	GetterLabel getters.Getter[string]
	// Icon represents the SVG icon name representing the trigger action (defaults to "plus").
	Icon string
	// IconClasses represents additional CSS classes applied specifically to the SVG icon wrapper.
	IconClasses string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this TableButtonCreate component.
func (e TableButtonCreate) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this TableButtonCreate.
func (e TableButtonCreate) GetRoles() []string {
	return e.Roles
}

// Build compiles the TableButtonCreate component into a standard ButtonLink node structure.
func (e TableButtonCreate) Build(ctx context.Context) gomponents.Node {
	icon := e.Icon
	if icon == "" {
		icon = tableButtonCreateDefaultIcon
	}
	classes := e.Classes
	if classes == "" {
		classes = tableButtonCreateDefaultClasses
	}
	var labelGetter getters.Getter[string]
	if e.GetterLabel != nil {
		labelGetter = e.GetterLabel
	} else if e.Label != "" {
		labelGetter = getters.Static(e.Label)
	}

	return ButtonLink{
		Page:        e.Page,
		Label:       labelGetter,
		Link:        e.Link,
		Icon:        icon,
		IconClasses: e.IconClasses,
		Classes:     classes,
	}.Build(ctx)
}
