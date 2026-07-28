package getters

// RowAttrNavigateFormat is like [Navigate] but returns nodes for [components.DataTable].RowAttr.
func RowAttrNavigateFormat(urlFormat string, g ...Getter[any]) Getter[map[string]string] {
	return rowAttrNavigateClick(Navigate(urlFormat, g...), nil)
}
