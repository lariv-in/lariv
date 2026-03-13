package p_courses

import (
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("courses.CourseMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Courses"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("All Courses"),
				Url:   lago.RoutePathGetter("courses.DefaultRoute"),
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Course: %s", getters.GetterKey("course.Name")),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Courses"),
			Url:   lago.RoutePathGetter("courses.DefaultRoute"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Course Detail"),
				Url:   getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("course.ID")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Course"),
				Url:   getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("course.ID")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Course"),
				Url:   getters.GetterFormat(AppUrl+"%v/delete/", getters.GetterKey("course.ID")),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("courses.CourseFilter", components.FormComponent{
		Url:    lago.RoutePathGetter("courses.DefaultRoute"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey("$get.Name")},
			components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey("$get.Code")},
			components.InputText{Label: "Subject", Name: "Subject", Getter: getters.GetterKey("$get.Subject")},
			components.InputText{Label: "Level", Name: "Level", Getter: getters.GetterKey("$get.Level")},
			components.InputTernary{
				Label:      "Active",
				Name:       "IsActive",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
				Getter:     getters.GetterKey("$get.IsActive"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseSelectionFilter", components.FormComponent{
		Url:    lago.RoutePathGetter("courses.SelectRoute"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey("$get.Name")},
			components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionFilter", components.FormComponent{
		Url:    lago.RoutePathGetter("courses.MultiSelectRoute"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey("$get.Name")},
			components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey("$get.Code")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})
}

// --- Form Fields & Forms ---

func courseFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					components.InputText{Label: "Course Name", Name: "Name", Required: true, Getter: getters.GetterKey("$in.Name")},
					components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey("$in.Code")},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					components.InputText{Label: "Subject", Name: "Subject", Getter: getters.GetterKey("$in.Subject")},
					components.InputText{Label: "Level", Name: "Level", Getter: getters.GetterKey("$in.Level")},
				},
			},
			components.InputTernary{
				Label:      "Active",
				Name:       "IsActive",
				TrueLabel:  "Yes",
				FalseLabel: "No",
				NoneLabel:  "Not Set",
				Getter:     getters.GetterKey("$in.IsActive"),
			},
			components.InputTextarea{
				Label:  "Description",
				Name:   "Description",
				Rows:   3,
				Getter: getters.GetterKey("$in.Description"),
			},
			components.ButtonSubmit{Label: "Save Course"},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("courses.CourseFormFields", courseFormFields())

	lago.RegistryPage.Register("courses.CourseCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:      lago.RoutePathGetter("courses.CreateRoute"),
				Method:   http.MethodPost,
				Title:    "Create Course",
				Subtitle: "Create a new course",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseFormFields(),
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Getter:   getters.GetterKey("course"),
				Url:      getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("$in.ID")),
				Method:   http.MethodPost,
				Title:    "Edit Course",
				Subtitle: "Update course details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseFormFields(),
				},
			},
		},
	})
}

// --- Table ---

func registerTablePages() {
	lago.RegistryPage.Register("courses.CourseTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseMenu"},
		},
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "course-table",
				Classes:         "w-full",
				Data:            getters.GetterKey("courses"),
				CreateUrl:       lago.RoutePathGetter("courses.CreateRoute"),
				OnClick:         getters.GetterNavigate(AppUrl+"%v/", getters.GetterKey("$row.ID")),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Code", Key: "Code", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Code")},
					}},
					{Label: "Subject", Key: "Subject", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Subject")},
					}},
					{Label: "Level", Key: "Level", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Level")},
					}},
					{Label: "Active", Key: "IsActive", Children: []components.PageInterface{
						components.FieldCheckbox{Getter: getters.GetterKey("$row.IsActive")},
					}},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("courses.CourseDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			components.Detail{
				Getter: getters.GetterKey("course"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Children: []components.PageInterface{
							components.FieldTitle{Getter: getters.GetterKey("$in.Name")},
							components.FieldSubtitle{Getter: getters.GetterKey("$in.Code")},
							components.LabelInline{
								Title:   "Subject",
								Classes: "mt-2",
								Children: []components.PageInterface{
									components.FieldText{Getter: getters.GetterKey("$in.Subject")},
								},
							},
							components.LabelInline{
								Title: "Level",
								Children: []components.PageInterface{
									components.FieldText{Getter: getters.GetterKey("$in.Level")},
								},
							},
							components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									components.FieldCheckbox{Getter: getters.GetterKey("$in.IsActive")},
								},
							},
							components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									components.FieldText{Getter: getters.GetterKey("$in.Description")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseDetailMenu"},
		},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this course?",
				CancelUrl: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("course.ID")),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("courses.CourseSelectionTable", components.Modal{
		UID:   "course-selection-modal",
		Title: "Select Course",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "course-selection-table",
				Data:            getters.GetterKey("courses"),
				OnClick:         getters.GetterSelect("course", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Code", Key: "Code", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Code")},
					}},
					{Label: "Level", Key: "Level", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Level")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseMultiSelectionTable", components.Modal{
		UID:   "course-multi-selection-modal",
		Title: "Select Courses",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "course-multi-selection-table",
				Data:            getters.GetterKey("courses"),
				OnClick:         getters.GetterMultiSelect("courses", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseMultiSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Code", Key: "Code", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Code")},
					}},
					{Label: "Level", Key: "Level", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Level")},
					}},
				},
			},
		},
	})
}

