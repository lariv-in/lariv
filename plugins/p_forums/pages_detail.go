package p_forums

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("forums.ForumThreadDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "forums.ForumThreadDetailMenu"}},
		Children: []components.PageInterface{
			&components.Detail[ForumThread]{
				Getter: getters.Key[ForumThread]("forum_thread"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "forums.ForumThreadDetailBody"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{Title: "Course", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Course.Code")},
							}},
							&components.LabelInline{Title: "Author", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Author.Name")},
							}},
							&components.LabelInline{Title: "Description", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$in.Description")},
							}},
							&components.LabelInline{Title: "Locked", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$in.Locked")},
							}},
						},
					},
				},
			},
		},
	})
	lago.RegistryPage.Register("forums.ForumThreadDeleteForm", &components.Modal{
		UID: "forum-thread-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{Title: "Confirm deletion", Message: "Delete this thread?", Attr: getters.FormBubbling(getters.Key[string]("$get.name"))},
		},
	})
}
