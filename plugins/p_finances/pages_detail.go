package p_finances

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("finances.StudentChargeDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "finances.StudentChargeDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[StudentCharge]{
				Getter: getters.Key[StudentCharge]("student_charge"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "finances.StudentChargeDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Description")},
							&components.LabelInline{Title: "Student", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Student.StudentNo")},
							}},
							&components.LabelInline{Title: "Semester", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Semester.Name")},
							}},
							&components.LabelInline{Title: "Amount (cents)", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int64]("$in.AmountCents")))},
							}},
							&components.LabelInline{Title: "Purpose", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Purpose")},
							}},
							&components.LabelInline{Title: "Remarks", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Remarks")},
							}},
							&components.LabelInline{Title: "Due on", Children: []components.PageInterface{
								&components.FieldDate{Getter: getters.Deref(getters.Key[*time.Time]("$in.DueOn"))},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("finances.StudentChargeDeleteForm", &components.Modal{
		UID: "student-charge-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this charge?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
