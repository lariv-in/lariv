package p_otp

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
)

const otpForgotPasswordLoginLinkKey = "otp.LoginForgotPasswordLink"

func patchUsersLoginPageWithOtpForgotLink(page components.PageInterface) components.PageInterface {
	scaffold, ok := page.(*components.ShellAuthScaffold)
	if !ok {
		panic("Base page for login page was not ShellAuthScaffold")
	}
	if otpForgotPasswordLinkPresent(scaffold) {
		return scaffold
	}

	col, ok := scaffold.Children[0].(*components.ContainerColumn)
	if !ok || len(col.Children) < 2 {
		panic("p_otp: unexpected login page layout (expect ContainerColumn with ≥2 children)")
	}
	formPost, ok := col.Children[1].(*components.FormListenBoostedPost)
	if !ok || len(formPost.Children) != 1 {
		panic("p_otp: unexpected login form wrapper layout")
	}
	fc, ok := formPost.Children[0].(*components.FormComponent[p_users.User])
	if !ok || fc.GetKey() != "p_users.AuthForm" {
		panic("p_otp: login FormListenBoostedPost must contain Auth FormComponent")
	}

	forgot := &components.ButtonLink{
		Page:  components.Page{Key: otpForgotPasswordLoginLinkKey},
		Label: getters.Static("Forgot password?"),
		Link:  lago.RoutePath("otp.ForgotPasswordRoute", nil),
	}

	newPost := *formPost
	newPost.Children = []components.PageInterface{fc, forgot}

	newCol := *col
	newCol.Children = append([]components.PageInterface(nil), col.Children...)
	newCol.Children[1] = &newPost

	newScaffold := *scaffold
	newScaffold.Children = []components.PageInterface{&newCol}

	return &newScaffold
}

func otpForgotPasswordLinkPresent(root components.ParentInterface) bool {
	for _, bl := range components.FindChildren[*components.ButtonLink](root) {
		if bl.GetKey() == otpForgotPasswordLoginLinkKey {
			return true
		}
	}
	return false
}

func pluginPages() lago.PluginFeatures[components.PageInterface] {
	auth := pageEntriesOtpAuth()
	prefs := pageEntriesOtpPreferences()
	entries := make([]registry.Pair[string, components.PageInterface], 0, len(auth)+len(prefs))
	entries = append(entries, auth...)
	entries = append(entries, prefs...)

	return lago.PluginFeatures[components.PageInterface]{
		Entries: entries,
		Patches: []registry.Pair[string, func(components.PageInterface) components.PageInterface]{
			{Key: "p_users.LoginPage", Value: patchUsersLoginPageWithOtpForgotLink},
		},
	}
}
