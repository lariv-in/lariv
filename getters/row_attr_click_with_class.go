package getters

import (
	"maragu.dev/gomponents"
)

// RowAttrClickWithClass merges an arbitrary @click expression with optional :class (e.g. filesystem selection).
func RowAttrClickWithClass(click, classExpr Getter[string]) Getter[gomponents.Node] {
	return rowAttrNavigateClick(click, classExpr)
}
