package getters

// RowAttrNavigate returns per-row attributes for HTMX navigation (list or grid styling from context).
func RowAttrNavigate[T comparable](urlGetter Getter[T]) Getter[map[string]string] {
	return rowAttrNavigateClick(NavigateGetter(urlGetter), nil)
}
