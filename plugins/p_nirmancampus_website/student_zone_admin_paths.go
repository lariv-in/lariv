package p_nirmancampus_website

import "github.com/lariv-in/lago/lago"

const StudentZoneAdminUrl = AppUrl + "student-zone/"

func init() {
	// --- Section routes ---
	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminDefaultRoute", lago.Route{
		Path:    StudentZoneAdminUrl,
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionListView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminSectionCreateRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "sections/create/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionCreateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminSectionDetailRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "sections/{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionDetailView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminSectionUpdateRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "sections/{id}/edit/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionUpdateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminSectionDeleteRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "sections/{id}/delete/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionDeleteView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminSectionSelectRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "sections/select/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminSectionSelectView"),
	})

	// --- Item routes ---
	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminItemListRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "items/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminItemListView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminItemCreateRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "items/create/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminItemCreateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminItemDetailRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "items/{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminItemDetailView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminItemUpdateRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "items/{id}/edit/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminItemUpdateView"),
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneAdminItemDeleteRoute", lago.Route{
		Path:    StudentZoneAdminUrl + "items/{id}/delete/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneAdminItemDeleteView"),
	})
}

