package p_nirmancampus_programs

import (
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// programsAdminRoleLayer limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var programsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

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
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.list", views.LayerList[Program]{
				Key:           getters.Static("programs"),
				QueryPatchers: programListQueryPatchers,
			}))

	lago.RegistryView.Register("programs.DetailView",
		lago.GetPageView("programs.ProgramDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.detail", views.LayerDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}))

	lago.RegistryView.Register("programs.CreateView",
		lago.GetPageView("programs.ProgramCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.admin_role", programsAdminRoleLayer).
			WithLayer("programs.create", views.LayerCreate[Program]{
				SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("programs.UpdateView",
		lago.GetPageView("programs.ProgramUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.admin_role", programsAdminRoleLayer).
			WithLayer("programs.detail", views.LayerDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}).
			WithLayer("programs.update", views.LayerUpdate[Program]{
				Key: getters.Static("program"),
				SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
					{Key: "programs.preload_structure_units", Value: queryPatcherPreloadProgramStructureUnits{}},
				},
			}))

	lago.RegistryView.Register("programs.DeleteView",
		lago.GetPageView("programs.ProgramDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.admin_role", programsAdminRoleLayer).
			WithLayer("programs.detail", views.LayerDetail[Program]{
				Key:          getters.Static("program"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
				},
			}).
			WithLayer("programs.delete", views.LayerDelete[Program]{
				Key:        getters.Static("program"),
				SuccessURL: lago.RoutePath("programs.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Program]{
					{Key: "programs.scope_by_role", Value: programScopeByRole{}},
				},
			}))

	lago.RegistryView.Register("programs.SelectView",
		lago.GetPageView("programs.ProgramSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.select_list", views.LayerList[Program]{
				Key:           getters.Static("programs"),
				QueryPatchers: programListQueryPatchers,
			}))

	lago.RegistryView.Register("programs.ProgramMediaMultiSelectView",
		lago.GetPageView("programs.ProgramMediaMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.program_media_multiselect_admin", programsAdminRoleLayer).
			WithLayer("programs.program_media_multiselect", views.LayerList[ProgramMedia]{
				Key: getters.Static("program_media"),
				QueryPatchers: views.QueryPatchers[ProgramMedia]{
					{Key: "programs.program_media_order", Value: queryPatcherProgramMediaOrder{}},
				},
			}))

	structureLayers := []struct {
		key string
		val views.Layer
	}{
		{"users.auth", p_users.AuthenticationLayer{}},
		{"programs.admin_role", programsAdminRoleLayer},
		{"programs.structure_load_program", programsStructureLoadProgramLayer{}},
	}
	applyStructure := func(v *views.View) *views.View {
		for _, m := range structureLayers {
			v = v.WithLayer(m.key, m.val)
		}
		return v
	}

	structureEditView := applyStructure(lago.GetPageView("programs.ProgramStructureEditPage")).
		WithLayer("programs.structure_edit_get", views.MethodLayer{
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
			WithLayer("programs.structure_unit_create", views.MethodLayer{
				Method: http.MethodPost,
				Handler: func(v *views.View) http.Handler {
					return handleStructureUnitCreate(v)
				},
			}))

	lago.RegistryView.Register("programs.StructureUnitEditModalView",
		applyStructure(lago.GetPageView("programs.StructureUnitEditModal")).
			WithLayer("programs.structure_unit_detail", views.LayerDetail[ProgramStructureUnit]{
				Key:          getters.Static("unit"),
				PathParamKey: getters.Static("unitId"),
				QueryPatchers: views.QueryPatchers[ProgramStructureUnit]{
					{Key: "programs.structure_unit_scope", Value: structureUnitScopeForContextProgram{}},
					{Key: "programs.structure_unit_preload_courses", Value: queryPatcherPreloadStructureUnitCourseAssociations{}},
				},
			}))

	lago.RegistryView.Register("programs.StructureUnitUpdateView",
		applyStructure(lago.GetPageView("programs.StructureUnitEditModal")).
			WithLayer("programs.structure_unit_detail", views.LayerDetail[ProgramStructureUnit]{
				Key:          getters.Static("unit"),
				PathParamKey: getters.Static("unitId"),
				QueryPatchers: views.QueryPatchers[ProgramStructureUnit]{
					{Key: "programs.structure_unit_scope", Value: structureUnitScopeForContextProgram{}},
					{Key: "programs.structure_unit_preload_courses", Value: queryPatcherPreloadStructureUnitCourseAssociations{}},
				},
			}).
			WithLayer("programs.structure_unit_update", views.MethodLayer{
				Method: http.MethodPost,
				Handler: func(v *views.View) http.Handler {
					return handleStructureUnitUpdate(v)
				},
			}))

	lago.RegistryView.Register("programs.StructureUnitDeleteView",
		applyStructure(lago.GetPageView("programs.StructureUnitDeleteForm")).
			WithLayer("programs.structure_unit_detail", views.LayerDetail[ProgramStructureUnit]{
				Key:          getters.Static("unit"),
				PathParamKey: getters.Static("unitId"),
				QueryPatchers: views.QueryPatchers[ProgramStructureUnit]{
					{Key: "programs.structure_unit_scope", Value: structureUnitScopeForContextProgram{}},
				},
			}).
			WithLayer("programs.structure_unit_delete", views.MethodLayer{
				Method: http.MethodPost,
				Handler: func(v *views.View) http.Handler {
					return handleStructureUnitDelete(v)
				},
			}))
}
