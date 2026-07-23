package p_blog

import (
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesSelection() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_blog.TagSelectionTable", Value: &components.Modal{
			UID: "tag-selection-modal",
			Children: []components.PageInterface{
				&components.DataTable[BlogTag]{
					UID:     "tag-selection-table",
					Title:   "Select Tags",
					Data:    getters.Key[components.ObjectList[BlogTag]]("tags"),
					RowAttr: getters.RowAttrSelectMulti(getters.IfOrElse(getters.Key[string]("$get.target_input"), getters.Static("Tags")), getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
					Columns: []components.TableColumn{
						{Label: "Name (ltree)", Name: "Name", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						}},
					},
				},
			},
		}},
	}
}
