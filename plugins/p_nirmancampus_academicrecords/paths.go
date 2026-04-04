package p_nirmancampus_academicrecords

import (
	"net/http"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
)

func registerRoutes() {
	_ = lago.RegistryRoute.Register("academicrecords.DefaultRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("academicrecords.ListView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.CreateRoute", lago.Route{
		Path:    AppUrl + "create/",
		Handler: lago.NewDynamicView("academicrecords.CreateView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.DetailRoute", lago.Route{
		Path:    AppUrl + "{id}/",
		Handler: lago.NewDynamicView("academicrecords.DetailView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.UpdateRoute", lago.Route{
		Path:    AppUrl + "{id}/edit/",
		Handler: lago.NewDynamicView("academicrecords.UpdateView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.DeleteRoute", lago.Route{
		Path:    AppUrl + "{id}/delete/",
		Handler: lago.NewDynamicView("academicrecords.DeleteView"),
	})

	_ = lago.RegistryRoute.Register("academicrecords.SelectRoute", lago.Route{
		Path:    AppUrl + "select/",
		Handler: lago.NewDynamicView("academicrecords.SelectView"),
	})
}

// legacyAcademicRecordsPathPrefix was the standalone app path before academic records moved under Students.
const legacyAcademicRecordsPathPrefix = "/academicrecords"

func legacyAcademicRecordsRedirect(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, legacyAcademicRecordsPathPrefix)
	if suffix == "" || suffix[0] != '/' {
		suffix = "/"
	}
	target := p_nirmancampus_students.AppUrl + "academicrecords" + suffix
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	lago.Redirect(w, r, target)
}

func registerLegacyAcademicRecordsRedirects() {
	h := http.HandlerFunc(legacyAcademicRecordsRedirect)
	p := legacyAcademicRecordsPathPrefix
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectList", lago.Route{Path: p + "/", Handler: h})
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectCreate", lago.Route{Path: p + "/create/", Handler: h})
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectSelect", lago.Route{Path: p + "/select/", Handler: h})
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectDetail", lago.Route{Path: p + "/{id}/", Handler: h})
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectUpdate", lago.Route{Path: p + "/{id}/edit/", Handler: h})
	_ = lago.RegistryRoute.Register("academicrecords.LegacyRedirectDelete", lago.Route{Path: p + "/{id}/delete/", Handler: h})
}

func init() {
	registerRoutes()
	registerLegacyAcademicRecordsRedirects()
}
