package p_nirmancampus_student_zone

import (
	"github.com/lariv-in/lago/lago"
)

func registerRoutes() {
	// --- Section routes ---
	_ = lago.RegistryRoute.Register("student_zone.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("student_zone.SectionListView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.SectionCreateRoute", lago.Route{
		Path:    AppUrl + "sections/create/",
		Handler: lago.NewDynamicView("student_zone.SectionCreateView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.SectionDetailRoute", lago.Route{
		Path:    AppUrl + "sections/{id}/",
		Handler: lago.NewDynamicView("student_zone.SectionDetailView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.SectionUpdateRoute", lago.Route{
		Path:    AppUrl + "sections/{id}/edit/",
		Handler: lago.NewDynamicView("student_zone.SectionUpdateView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.SectionDeleteRoute", lago.Route{
		Path:    AppUrl + "sections/{id}/delete/",
		Handler: lago.NewDynamicView("student_zone.SectionDeleteView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.SectionSelectRoute", lago.Route{
		Path:    AppUrl + "sections/select/",
		Handler: lago.NewDynamicView("student_zone.SectionSelectView"),
	})

	// --- Item routes ---
	_ = lago.RegistryRoute.Register("student_zone.ItemListRoute", lago.Route{
		Path:    AppUrl + "items/",
		Handler: lago.NewDynamicView("student_zone.ItemListView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.ItemCreateRoute", lago.Route{
		Path:    AppUrl + "items/create/",
		Handler: lago.NewDynamicView("student_zone.ItemCreateView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.ItemDetailRoute", lago.Route{
		Path:    AppUrl + "items/{id}/",
		Handler: lago.NewDynamicView("student_zone.ItemDetailView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.ItemUpdateRoute", lago.Route{
		Path:    AppUrl + "items/{id}/edit/",
		Handler: lago.NewDynamicView("student_zone.ItemUpdateView"),
	})

	_ = lago.RegistryRoute.Register("student_zone.ItemDeleteRoute", lago.Route{
		Path:    AppUrl + "items/{id}/delete/",
		Handler: lago.NewDynamicView("student_zone.ItemDeleteView"),
	})
}

func init() {
	registerRoutes()
}
