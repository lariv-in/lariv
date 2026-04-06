package p_filesystem

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func selectionTable(name, filterName, childRoute string, multi, selectDirectories bool) *components.Modal {
	title := "Select Directory"
	subtitle := "Choose a folder"
	if !selectDirectories {
		title = "Select Files"
		subtitle = "Choose files, or open folders to browse deeper"
	}

	modalID := "filesystem-selection-modal-" + name
	onClick := selectionRowClickGetter(name, modalID, childRoute, multi, selectDirectories)
	var rowClass getters.Getter[string]
	if multi {
		rowClass = getters.SelectMultiRowClass(selectionTargetInput(name), getters.Key[uint]("$row.ID"))
	}

	return &components.Modal{
		UID: modalID,
		Children: []components.PageInterface{
			&components.DataTable[VNode]{
				UID:      "filesystem-selection-table-" + name,
				Title:    title,
				Subtitle: subtitle,
				Data:     getters.Key[components.ObjectList[VNode]]("vnodes"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: filterName}},
				},
				RowAttr: getters.RowAttrClickWithClass(onClick, rowClass),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Path", Name: "Path", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")}}},
				},
			},
		},
	}
}

func registerSelection() {
	lago.RegistryPage.Register("filesystem.ParentSelectionTable", selectionTable("ParentID", "filesystem.ParentSelectionFilter", "filesystem.SelectChildRoute", false, true))
	lago.RegistryPage.Register("filesystem.MultiSelectionTable", selectionTable("ParentID", "filesystem.ParentSelectionFilter", "filesystem.MultiSelectChildRoute", true, false))
	lago.RegistryPage.Register("filesystem.DestinationSelectionTable", selectionTable("DestinationID", "filesystem.DestinationSelectionFilter", "filesystem.MoveSelectChildRoute", false, true))
}

func registerDelete() {
	lago.RegistryPage.Register("filesystem.VNodeDeleteForm", &components.Modal{
		UID: "filesystem-vnode-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this item? Deleting directories will remove all nested contents.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
