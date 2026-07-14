package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
)

// GetterPage represents a component that resolves another page/component dynamically at render time.
// It uses a Getter to lookup the target component from the context or configuration registry, then renders it.
//
// Use Cases:
//   - Dynamically rendering distinct sub-panels or widget dashboards based on user preferences, feature flags, or status flags.
//
// Example:
//
//	// Supposing we have a component registry:
//	// var RegistryWidgets = registry.NewRegistry[components.PageInterface]()
//
//	&components.GetterPage{
//	    Getter: RegistryWidgets.Getter("user_analytics_card"),
//	}
type GetterPage struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the PageInterface component to render.
	Getter getters.Getter[PageInterface]
}

// Build compiles the GetterPage by resolving the dynamically retrieved page component and rendering it.
func (e GetterPage) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	page, err := e.Getter(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if page == nil {
		return Group{}
	}
	return Render(page, ctx)
}

// GetKey returns the unique key identifier for this GetterPage component.
func (e GetterPage) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this GetterPage.
func (e GetterPage) GetRoles() []string {
	return e.Roles
}
