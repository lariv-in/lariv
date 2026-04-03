package views

import (
	"maps"
	"net/http"

	"github.com/lariv-in/lago/registry"
)

type FormPatcher interface {
	Patch(view View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error)
}

type FormPatchers []registry.Pair[string, FormPatcher]

func (f FormPatchers) Apply(view View, r *http.Request, values map[string]any, errors map[string]error) (map[string]any, map[string]error) {
	for _, formPatcher := range f {
		newValues, newErrors := formPatcher.Value.Patch(view, r, values, errors)
		maps.Copy(values, newValues)
		maps.Copy(errors, newErrors)
	}
	return values, errors
}
