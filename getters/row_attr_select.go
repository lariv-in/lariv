package getters

import (
	"maragu.dev/gomponents"
)

// RowAttrSelect returns per-row attributes for single foreign-key selection rows.
func RowAttrSelect[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[gomponents.Node] {
	return rowAttrNavigateClick(Select(name, valueGetter, displayGetter), nil)
}
