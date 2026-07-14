package components

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Detail binds a generic model object of type T resolved from the context to its child components under the context key "$in".
// Child components (e.g., fields) can then resolve their values using getters relative to "$in" (e.g., getters.Key("$in.FieldName")).
//
// Use Cases:
//   - Displaying read-only fields of a database entity (e.g., User profiles, Invoice details, Product descriptions).
//   - Scoping complex sub-components to specific nested data models.
//
// Example:
//
//	&components.Detail[User]{
//	    Getter: getters.Key[User]("active_user"),
//	    Children: []components.PageInterface{
//	        &components.FieldText{Getter: getters.Key[string]("$in.Name")},
//	        &components.FieldText{Getter: getters.Key[string]("$in.Email")},
//	    },
//	}
type Detail[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the entity/object of type T.
	Getter getters.Getter[T]
	// Children represents the components nested inside the detail layout scope.
	Children []PageInterface
}

// Build compiles the Detail component, binding the model object to context and rendering its children.
func (e Detail[T]) Build(ctx context.Context) Node {
	childCtx := ctx
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("Detail getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if v := reflect.ValueOf(value); v.IsValid() && !v.IsZero() {
			objMap := getters.MapFromStruct(value)
			childCtx = context.WithValue(ctx, "$in", objMap)
		}
	}

	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, childCtx))
	}
	return Div(Group(childNodes))
}

// GetKey returns the unique key identifier for this Detail component.
func (e Detail[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this Detail.
func (e Detail[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components nested inside the detail layout scope.
func (e Detail[T]) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components nested inside the detail layout scope.
func (e *Detail[T]) SetChildren(children []PageInterface) {
	e.Children = children
}
