package views

import (
	"maps"
	"net/http"

	"github.com/lariv-in/lariv/registry"
)

// FormPatcher defines an interface for components capable of modifying form values and validation errors.
// It is applied during form submission flows to run hooks, inject defaults, or inject custom validations.
//
// Use Cases:
//   - Injecting session data (e.g. current user ID) into submitted form maps.
//   - Performing cross-field validations and appending custom validation errors.
//   - Hashing sensitive inputs (e.g. passwords) before storage persistence.
//
// Example:
//
//	type AuthorPatcher struct{}
//
//	func (p AuthorPatcher) Patch(view views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
//		formData["author_id"] = GetUserIDFromSession(r)
//		return formData, formErrors
//	}
type FormPatcher interface {
	// Patch performs modifications on form data values and validation errors, returning the altered maps.
	Patch(view View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error)
}

// FormPatchers represents an ordered sequence of FormPatcher registry pairs.
type FormPatchers []registry.Pair[string, FormPatcher]

// Apply executes all nested [FormPatcher] components sequentially, merging their output values and errors.
func (f FormPatchers) Apply(view View, r *http.Request, values map[string]any, errors map[string]error) (map[string]any, map[string]error) {
	for _, formPatcher := range f {
		newValues, newErrors := formPatcher.Value.Patch(view, r, values, errors)
		maps.Copy(values, newValues)
		maps.Copy(errors, newErrors)
	}
	return values, errors
}
