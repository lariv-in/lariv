package getters

// RowAttrClickWithClass merges an arbitrary @click expression with optional :class (e.g. filesystem selection).
func RowAttrClickWithClass(click, classExpr Getter[string]) Getter[map[string]string] {
	return rowAttrNavigateClick(click, classExpr)
}
