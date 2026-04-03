package p_nirmancampus_programs

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// programsAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var programsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware{Roles: []string{"admin"}}

func init() {
	univPatcher := queryPatcherUniversity{Param: "University"}
	programTypePatcher := queryPatcherProgramType{Param: "ProgramType"}

	programListQueryPatchers := views.QueryPatchers[Program]{
		{Key: "nirmancampus_programs.filter_university", Value: univPatcher},
		{Key: "nirmancampus_programs.filter_program_type", Value: programTypePatcher},
		{Key: "programs.scope_by_role", Value: programScopeByRole{}},
	}

	lago.RegistryView.Register("programs.ListView",
		lago.GetPageView("programs.ProgramTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.list", views.MiddlewareList[Program]{
				Key:           getters.Static("programs"),
				QueryPatchers: programListQueryPatchers,
			}))

	lago.RegistryView.Register("programs.DetailView",
		lago.GetPageView("programs.ProgramDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.detail", views.MiddlewareDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}))

	lago.RegistryView.Register("programs.CreateView",
		lago.GetPageView("programs.ProgramCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithMiddleware("programs.create", views.MiddlewareCreate[Program]{
				SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("programs.UpdateView",
		lago.GetPageView("programs.ProgramUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithMiddleware("programs.detail", views.MiddlewareDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}).
			WithMiddleware("programs.update", views.MiddlewareUpdate[Program]{
				Key: getters.Static("program"),
				SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}))

	lago.RegistryView.Register("programs.DeleteView",
		lago.GetPageView("programs.ProgramDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.admin_role", programsAdminRoleMiddleware).
			WithMiddleware("programs.detail", views.MiddlewareDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
				},
			}).
			WithMiddleware("programs.delete", views.MiddlewareDelete[Program]{
				Key:        getters.Static("program"),
				SuccessURL: lago.RoutePath("programs.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
				},
			}))

	lago.RegistryView.Register("programs.SelectView",
		lago.GetPageView("programs.ProgramSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("programs.select_list", views.MiddlewareList[Program]{
				Key:           getters.Static("programs"),
				QueryPatchers: programListQueryPatchers,
			}))

	structureMiddlewares := []struct {
		key string
		val views.Middleware
	}{
		{"users.auth", p_users.AuthenticationMiddleware{}},
		{"programs.admin_role", programsAdminRoleMiddleware},
		{"programs.structure_load_program", programsStructureLoadProgramMiddleware{}},
	}
	applyStructure := func(v *views.View) *views.View {
		for _, m := range structureMiddlewares {
			v = v.WithMiddleware(m.key, m.val)
		}
		return v
	}

	structureEditView := applyStructure(lago.GetPageView("programs.ProgramStructureEditPage")).
		WithMiddleware("programs.structure_edit_get", views.MethodMiddleware{
			Method: http.MethodGet,
			Handler: func(v *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					v.RenderPage(w, r)
				})
			},
		})
	lago.RegistryView.Register("programs.StructureEditView", structureEditView)

	lago.RegistryView.Register("programs.StructureUnitCreateModalView",
		applyStructure(lago.GetPageView("programs.StructureUnitCreateModal")))

	lago.RegistryView.Register("programs.StructureUnitCreateView",
		applyStructure(lago.GetPageView("programs.StructureUnitCreateModal")).
			WithMiddleware("programs.structure_unit_create", views.MethodMiddleware{
				Method: http.MethodPost,
				Handler: func(v *views.View) http.Handler {
					return handleStructureUnitCreate(v)
				},
			}))

	lago.RegistryView.Register("programs.StructureUnitEditModalView",
		applyStructure(lago.GetPageView("programs.StructureUnitEditModal")).
			WithMiddleware("programs.structure_unit_detail", views.MiddlewareDetail[ProgramStructureUnit]{
				Key:          getters.Static("unit"),
				PathParamKey: getters.Static("unitId"),
				QueryPatchers: views.QueryPatchers[ProgramStructureUnit]{
					{Key: "programs.structure_unit_scope", Value: structureUnitScopeForContextProgram{}},
					{Key: "programs.structure_unit_preload_courses", Value: queryPatcherPreloadStructureUnitCourseAssociations{}},
				},
			}))

	lago.RegistryView.Register("programs.StructureUnitUpdateView",
		applyStructure(lago.GetPageView("programs.StructureUnitEditModal")).
			WithMiddleware("programs.structure_unit_detail", views.MiddlewareDetail[ProgramStructureUnit]{
				Key:          getters.Static("unit"),
				PathParamKey: getters.Static("unitId"),
				QueryPatchers: views.QueryPatchers[ProgramStructureUnit]{
					{Key: "programs.structure_unit_scope", Value: structureUnitScopeForContextProgram{}},
					{Key: "programs.structure_unit_preload_courses", Value: queryPatcherPreloadStructureUnitCourseAssociations{}},
				},
			}).
			WithMiddleware("programs.structure_unit_update", views.MethodMiddleware{
				Method: http.MethodPost,
				Handler: func(v *views.View) http.Handler {
					return handleStructureUnitUpdate(v)
				},
			}))
}
