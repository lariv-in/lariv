package p_forums

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_users"
)

func forumThreadFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "forums.ForumThreadFormFields"},
		Children: []components.PageInterface{
			&components.InputForeignKey[p_courses.Course]{Label: "Course", Name: "CourseID", Required: true, Url: lago.RoutePath("courses.SelectRoute", nil), Display: getters.Key[string]("$in.Code"), Placeholder: "Select course...", Getter: getters.Association[p_courses.Course](getters.Key[uint]("$in.CourseID"))},
			&components.InputText{Label: "Title", Name: "Title", Required: true, Getter: getters.Key[string]("$in.Title")},
			&components.InputTextarea{Label: "Description", Name: "Description", Rows: 6, Getter: getters.Key[string]("$in.Description")},
			&components.InputForeignKey[p_users.User]{Label: "Author (optional)", Name: "UserID", Required: false, Url: lago.RoutePath("users.SelectRoute", nil), Display: getters.Key[string]("$in.Name"), Placeholder: "User…", Getter: getters.Association[p_users.User](getters.Deref(getters.Key[*uint]("$in.UserID")))},
			&components.InputCheckbox{Label: "Locked", Name: "Locked", Getter: getters.Key[bool]("$in.Locked")},
		},
	}
}

func registerFormPages() {
	dn := getters.Static("forums.ForumThreadDeleteForm")
	lago.RegistryPage.Register("forums.ForumThreadCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "forums.ForumThreadMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("forums.ForumThreadCreateForm"),
				ActionURL: lago.RoutePath("forums.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ForumThread]{
						Attr:           getters.FormBubbling(getters.Static("forums.ForumThreadCreateForm")),
						Title:          "New thread",
						ChildrenInput:  []components.PageInterface{forumThreadFormFields()},
						ChildrenAction: []components.PageInterface{&components.ButtonSubmit{Label: "Save"}},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("forums.ForumThreadUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "forums.ForumThreadDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("forums.ForumThreadUpdateForm"),
				ActionURL: lago.RoutePath("forums.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[ForumThread]{
						Getter:        getters.Key[ForumThread]("forum_thread"),
						Attr:          getters.FormBubbling(getters.Static("forums.ForumThreadUpdateForm")),
						Title:         "Edit thread",
						ChildrenInput: []components.PageInterface{forumThreadFormFields()},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex gap-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label: "Delete", Icon: "trash", Name: dn,
										Url:         lago.RoutePath("forums.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))}),
										FormPostURL: lago.RoutePath("forums.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("forum_thread.ID"))}),
										ModalUID:    "forum-thread-delete-modal", Classes: "btn-error",
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
