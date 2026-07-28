package components

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"slices"
)

// PageInterface represents the standard component interface within the Lariv UI framework.
// Every custom page layout or interactive component must satisfy this interface.
type PageInterface interface {
	// Build compiles the component using context values and writes HTML to w.
	// cat is the compiled app catalog (pages, shell chrome); leaf components may ignore it.
	Build(cat Catalog, ctx context.Context, w io.Writer) error
	// GetKey returns the unique key identifying this specific component.
	GetKey() string
	// GetRoles returns the authorized roles allowed to view or interact with this component.
	GetRoles() []string
}

// Page struct defines common properties embedded in all component structs.
// It carries the unique component key and routing roles configuration.
type Page struct {
	// Key represents the unique component key identifier.
	Key string
	// Roles represents a slice of authorized roles required to view this component.
	Roles []string
}

// GetKey returns the unique key identifier for this Page.
func (p Page) GetKey() string {
	return p.Key
}

// GetRoles returns the authorized roles required to view this Page.
func (p Page) GetRoles() []string {
	return p.Roles
}

// Render compiles the page component if the role in ctx (under key "$role") matches the required roles.
// If the user's role is unauthorized, it writes nothing.
func Render(p PageInterface, cat Catalog, ctx context.Context, w io.Writer) error {
	if p == nil {
		return nil
	}
	if cat == nil {
		cat = EmptyCatalog{}
	}
	roles := GetRequiredRoles(p)
	if roles == nil {
		return p.Build(cat, ctx, w)
	}
	currentRole, _ := ctx.Value("$role").(string)
	if slices.Contains(roles, currentRole) {
		return p.Build(cat, ctx, w)
	}
	return nil
}

// RenderHTML renders a page component to an HTML fragment.
func RenderHTML(p PageInterface, cat Catalog, ctx context.Context) (template.HTML, error) {
	var b bytes.Buffer
	if err := Render(p, cat, ctx, &b); err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}

// GetRequiredRoles extracts the required roles list configured in the component's embedded Page structure.
func GetRequiredRoles(p PageInterface) []string {
	return p.GetRoles()
}
