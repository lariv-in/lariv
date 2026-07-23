package p_no_signup

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
)

func patchUsersLoginPageRemoveSignup(page components.PageInterface) components.PageInterface {
	scaffold, ok := page.(*components.ShellAuthScaffold)
	if !ok {
		return page
	}
	if len(scaffold.Children) == 0 {
		return scaffold
	}
	col, ok := scaffold.Children[0].(*components.ContainerColumn)
	if !ok || len(col.Children) < 2 {
		return scaffold
	}
	formPost, ok := col.Children[1].(*components.FormListenBoostedPost)
	if !ok || len(formPost.Children) != 1 {
		return scaffold
	}
	fc, ok := formPost.Children[0].(*components.FormComponent[p_users.User])
	if !ok {
		return scaffold
	}

	var newAction []components.PageInterface
	for _, act := range fc.ChildrenAction {
		if act.GetKey() != "p_users.AuthSignupLink" {
			newAction = append(newAction, act)
		}
	}

	newFc := *fc
	newFc.ChildrenAction = newAction

	newPost := *formPost
	newPost.Children = []components.PageInterface{&newFc}

	newCol := *col
	newCol.Children = append([]components.PageInterface(nil), col.Children...)
	newCol.Children[1] = &newPost

	newScaffold := *scaffold
	newScaffold.Children = []components.PageInterface{&newCol}

	return &newScaffold
}

func pluginPages() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{
		Patches: []registry.Pair[string, func(components.PageInterface) components.PageInterface]{
			{Key: "p_users.LoginPage", Value: patchUsersLoginPageRemoveSignup},
		},
	}
}
