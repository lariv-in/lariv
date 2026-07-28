package getters

// RowAttrSelect returns per-row attributes for single foreign-key selection rows.
func RowAttrSelect[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[map[string]string] {
	return rowAttrNavigateClick(Select(name, valueGetter, displayGetter), nil)
}

// RowAttrSelectNamed is like [RowAttrSelect] but resolves the input name at render time from nameGetter
// (use with target_input on the picker URL so each row in a dynamic list gets a distinct fk-select target).
func RowAttrSelectNamed[T, D comparable](nameGetter Getter[string], valueGetter Getter[T], displayGetter Getter[D]) Getter[map[string]string] {
	return rowAttrNavigateClick(SelectNamed(nameGetter, valueGetter, displayGetter), nil)
}
