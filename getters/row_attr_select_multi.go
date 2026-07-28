package getters

// RowAttrSelectMulti combines multi-select click + row class for m2m selector tables.
func RowAttrSelectMulti[T, D comparable](nameGetter Getter[string], valueGetter Getter[T], displayGetter Getter[D]) Getter[map[string]string] {
	return rowAttrNavigateClick(
		SelectMulti(nameGetter, valueGetter, displayGetter),
		SelectMultiRowClass(nameGetter, valueGetter),
	)
}
