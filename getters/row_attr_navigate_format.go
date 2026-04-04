package getters

import (
	"maragu.dev/gomponents"
)

// RowAttrNavigateFormat is like [Navigate] but returns nodes for [components.DataTable].RowAttr.
func RowAttrNavigateFormat(urlFormat string, g ...Getter[any]) Getter[gomponents.Node] {
	return rowAttrNavigateClick(Navigate(urlFormat, g...), nil)
}
