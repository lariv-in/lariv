package p_nirmancampus_programs

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// programsAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var programsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func init() {
	univPatcher := queryPatcherUniversity("University")
	programTypePatcher := queryPatcherProgramType("ProgramType")

	lago.RegistryView.Register("programs.ListView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("nirmancampus_programs.filter_university", univPatcher).
			WithQueryPatcher("nirmancampus_programs.filter_program_type", programTypePatcher).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole))

	lago.RegistryView.Register("programs.DetailView",
		views.DetailView[Program]("program")(
			lago.GetPageView("programs.ProgramDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole))

	lago.RegistryView.Register("programs.CreateView",
		views.CreateView[Program](
			lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("programs.ProgramCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware))

	lago.RegistryView.Register("programs.UpdateView",
		views.DetailView[Program]("program")(
			views.UpdateView[Program](
				lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("programs.ProgramUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole))

	lago.RegistryView.Register("programs.DeleteView",
		views.DetailView[Program]("program")(
			views.DeleteView[Program](
				lago.GetterRoutePath("programs.DefaultRoute", nil),
			)(
				lago.GetPageView("programs.ProgramDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole))

	lago.RegistryView.Register("programs.SelectView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("nirmancampus_programs.filter_university", univPatcher).
			WithQueryPatcher("nirmancampus_programs.filter_program_type", programTypePatcher).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole))
}
