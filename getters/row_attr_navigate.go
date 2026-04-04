package getters

import (
	"maragu.dev/gomponents"
)

// RowAttrNavigate returns per-row attributes for HTMX navigation (list or grid styling from context).
func RowAttrNavigate[T comparable](urlGetter Getter[T]) Getter[gomponents.Node] {
	return rowAttrNavigateClick(NavigateGetter(urlGetter), nil)
}
