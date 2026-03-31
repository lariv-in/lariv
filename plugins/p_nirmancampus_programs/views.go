package p_nirmancampus_programs

import (
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func programPageLookup(name string) (components.PageInterface, bool) {
	return lago.RegistryPage.Get(name)
}

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
		views.DetailView[Program]("program", "id")(
			lago.GetPageView("programs.ProgramDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole).
			WithQueryPatcher("programs.preload_structure_units", queryPatcherPreloadProgramStructureUnits()))

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
		views.DetailView[Program]("program", "id")(
			views.UpdateView[Program]("id",
				lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("programs.ProgramUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithQueryPatcher("programs.scope_by_role", ProgramScopeByRole).
			WithQueryPatcher("programs.preload_structure_units", queryPatcherPreloadProgramStructureUnits()))

	lago.RegistryView.Register("programs.DeleteView",
		views.DetailView[Program]("program", "id")(
			views.DeleteView[Program]("id",
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

	structureMiddlewares := []struct {
		key string
		fn  views.Middleware
	}{
		{"users.auth", p_users.AuthenticationMiddleware},
		{"programs.admin_role", programsAdminRoleMiddleware},
		{"programs.structure_load_program", middlewareProgramsStructureLoadProgram},
	}
	applyStructureMiddlewares := func(v *views.View) *views.View {
		for _, m := range structureMiddlewares {
			v = v.WithMiddleware(m.key, m.fn)
		}
		return v
	}

	structureEditView := &views.View{
		PageName: "programs.ProgramStructureEditPage",
		PageLookup: programPageLookup,
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: func(v *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					v.ServeRenderPage(w, r)
				})
			},
		},
	}
	lago.RegistryView.Register("programs.StructureEditView", applyStructureMiddlewares(structureEditView))

	structureUnitCreateModalView := applyStructureMiddlewares(lago.GetPageView("programs.StructureUnitCreateModal"))
	lago.RegistryView.Register("programs.StructureUnitCreateModalView", structureUnitCreateModalView)

	structureUnitCreatePost := &views.View{
		PageName:   "programs.StructureUnitCreateModal",
		PageLookup: programPageLookup,
		Handlers:   map[string]func(*views.View) http.Handler{},
	}
	structureUnitCreatePost = structureUnitCreatePost.
		WithFormPatcher("programs.structure_unit_program_id", formPatcherStructureUnitProgramIDFromPath).
		WithMethod(http.MethodPost, func(v *views.View) http.Handler {
			return handleStructureUnitCreate(v)
		})
	lago.RegistryView.Register("programs.StructureUnitCreateView", applyStructureMiddlewares(structureUnitCreatePost))

	structureUnitEditModalView := views.DetailView[ProgramStructureUnit]("unit", "unitId")(
		lago.GetPageView("programs.StructureUnitEditModal"),
	)
	structureUnitEditModalView = structureUnitEditModalView.WithQueryPatcher(
		"programs.structure_unit_scope", queryPatcherStructureUnitForContextProgram).
		WithQueryPatcher("programs.structure_unit_preload_courses", queryPatcherPreloadStructureUnitCourseAssociations())
	lago.RegistryView.Register("programs.StructureUnitEditModalView", applyStructureMiddlewares(structureUnitEditModalView))

	structureUnitUpdatePost := &views.View{
		PageName:   "programs.StructureUnitEditModal",
		PageLookup: programPageLookup,
		Handlers:   map[string]func(*views.View) http.Handler{},
	}
	structureUnitUpdatePost = structureUnitUpdatePost.
		WithFormPatcher("programs.structure_unit_program_id", formPatcherStructureUnitProgramIDFromPath).
		WithMethod(http.MethodPost, func(v *views.View) http.Handler {
			return handleStructureUnitUpdate(v)
		})
	lago.RegistryView.Register("programs.StructureUnitUpdateView", applyStructureMiddlewares(structureUnitUpdatePost))
}
