package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// websiteIntelContextLayer loads intel flags and detail href into context after [views.LayerDetail] for [Website].
type websiteIntelContextLayer struct{}

func (websiteIntelContextLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		site, ok := ctx.Value("website").(Website)
		if !ok {
			site = Website{}
		}
		setEmpty := func() {
			ctx = context.WithValue(ctx, "websiteIntelAddVisible", false)
			ctx = context.WithValue(ctx, "websiteIntelLinkVisible", false)
			ctx = context.WithValue(ctx, "websiteIntelDetailHref", "")
		}
		if site.ID == 0 {
			setEmpty()
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("seer_websites: website intel context: db", "error", err)
			setEmpty()
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		exists, err := p_seer_intel.IntelExistsForSource(ctx, db, (Website{}).Kind(), site.ID)
		if err != nil {
			slog.Error("seer_websites: website intel context: exists check", "error", err)
			exists = false
		}
		href := ""
		if exists {
			href, err = p_seer_intel.IntelDetailPathForSource(ctx, (Website{}).Kind(), site.ID)
			if err != nil {
				slog.Error("seer_websites: website intel context: detail path", "error", err)
				href = ""
			}
		}
		ctx = context.WithValue(ctx, "websiteIntelAddVisible", !exists)
		ctx = context.WithValue(ctx, "websiteIntelLinkVisible", exists)
		ctx = context.WithValue(ctx, "websiteIntelDetailHref", href)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// websiteRunnerWorkerPoolStateLayer sets [workerPoolIsRunning] from in-process pool state after [views.LayerDetail] for [WebsiteRunner].
type websiteRunnerWorkerPoolStateLayer struct{}

func (websiteRunnerWorkerPoolStateLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		run, ok := ctx.Value("websiteRunner").(WebsiteRunner)
		running := ok && run.ID != 0 && WebsiteRunnerWorkerPoolIsRunning(run.ID)
		ctx = context.WithValue(ctx, "workerPoolIsRunning", running)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	websitePatchers := views.QueryPatchers[Website]{
		{Key: "seer_websites.website.not_deleted", Value: websiteActiveOnlyPatcher{}},
		{Key: "seer_websites.website.order", Value: views.QueryPatcherOrderBy[Website]{Order: "id DESC"}},
	}

	websiteDetailPatchers := views.QueryPatchers[Website]{
		{Key: "seer_websites.website_detail.not_deleted", Value: websiteActiveOnlyPatcher{}},
	}

	lago.RegistryView.Register("seer_websites.WebsiteListView",
		lago.GetPageView("seer_websites.WebsiteTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.list", views.LayerList[Website]{
				Key:           getters.Static("websites"),
				QueryPatchers: websitePatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteAddView",
		lago.GetPageView("seer_websites.WebsiteAddForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.create", views.LayerCreate[Website]{
				SuccessURL: lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website.scrape", Value: websiteScrapeFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteDetailView",
		lago.GetPageView("seer_websites.WebsiteDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}).
			WithLayer("seer_websites.website.intel", websiteIntelContextLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteAddIntelView",
		lago.GetPageView("seer_websites.WebsiteDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.add_intel_detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}).
			WithLayer("seer_websites.website.add_intel", websiteAddIntelLayer{}).
			WithLayer("seer_websites.website.intel", websiteIntelContextLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteAddAllIntelView",
		lago.GetPageView("seer_websites.WebsiteTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.add_all_intel", websiteAddAllIntelLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteSoftDeleteView",
		lago.GetPageView("seer_websites.WebsiteDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website.delete_detail", views.LayerDetail[Website]{
				Key:           getters.Static("website"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteDetailPatchers,
			}).
			WithLayer("seer_websites.website.soft_delete", websiteSoftDeleteLayer{}))

	websiteSourcePatchers := views.QueryPatchers[WebsiteSource]{
		{Key: "seer_websites.website_source.order", Value: views.QueryPatcherOrderBy[WebsiteSource]{Order: "id DESC"}},
	}

	websiteSourceDetailPatchers := views.QueryPatchers[WebsiteSource]{
		{Key: "seer_websites.website_source.preload_runner", Value: views.QueryPatcherPreload[WebsiteSource]{Fields: []string{"WebsiteRunner"}}},
	}

	runnerPatchers := views.QueryPatchers[WebsiteRunner]{
		{Key: "seer_websites.website_runner.order", Value: views.QueryPatcherOrderBy[WebsiteRunner]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("seer_websites.WebsiteSourceListView",
		lago.GetPageView("seer_websites.WebsiteSourceTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.list", views.LayerList[WebsiteSource]{
				Key:           getters.Static("websiteSources"),
				QueryPatchers: websiteSourcePatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteSourceDetailView",
		lago.GetPageView("seer_websites.WebsiteSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.detail", views.LayerDetail[WebsiteSource]{
				Key:           getters.Static("websiteSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteSourceDetailPatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteSourceCreateView",
		lago.GetPageView("seer_websites.WebsiteSourceCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.create", views.LayerCreate[WebsiteSource]{
				SuccessURL: lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website_source.validate", Value: websiteSourceValidate{}},
					{Key: "seer_websites.website_source.url_pageurl", Value: websiteSourcePageURLFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteSourceUpdateView",
		lago.GetPageView("seer_websites.WebsiteSourceUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.detail_for_update", views.LayerDetail[WebsiteSource]{
				Key:           getters.Static("websiteSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteSourceDetailPatchers,
			}).
			WithLayer("seer_websites.website_source.update", views.LayerUpdate[WebsiteSource]{
				Key: getters.Static("websiteSource"),
				SuccessURL: lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteSource.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website_source.validate", Value: websiteSourceValidate{}},
					{Key: "seer_websites.website_source.url_pageurl", Value: websiteSourcePageURLFormPatcher{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteSourceDeleteView",
		lago.GetPageView("seer_websites.WebsiteSourceDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.delete_detail", views.LayerDetail[WebsiteSource]{
				Key:           getters.Static("websiteSource"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: websiteSourceDetailPatchers,
			}).
			WithLayer("seer_websites.website_source.delete", views.LayerDelete[WebsiteSource]{
				Key:        getters.Static("websiteSource"),
				SuccessURL: lago.RoutePath("seer_websites.WebsiteSourceListRoute", nil),
			}))

	lago.RegistryView.Register("seer_websites.WebsiteSourceFetchView",
		lago.GetPageView("seer_websites.WebsiteSourceDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_source.fetch_detail", views.LayerDetail[WebsiteSource]{
				Key:           getters.Static("websiteSource"),
				PathParamKey:  getters.Static("source_id"),
				QueryPatchers: websiteSourceDetailPatchers,
			}).
			WithLayer("seer_websites.website_source.fetch_action", websiteSourceFetchActionLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerListView",
		lago.GetPageView("seer_websites.WebsiteRunnerTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.list", views.LayerList[WebsiteRunner]{
				Key:           getters.Static("websiteRunners"),
				QueryPatchers: runnerPatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerSelectView",
		lago.GetPageView("seer_websites.WebsiteRunnerSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.select_list", views.LayerList[WebsiteRunner]{
				Key:           getters.Static("websiteRunners"),
				QueryPatchers: runnerPatchers,
			}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerDetailView",
		lago.GetPageView("seer_websites.WebsiteRunnerDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.detail", views.LayerDetail[WebsiteRunner]{
				Key:          getters.Static("websiteRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_websites.website_runner.worker_pool_state", websiteRunnerWorkerPoolStateLayer{}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerCreateView",
		lago.GetPageView("seer_websites.WebsiteRunnerCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.create", views.LayerCreate[WebsiteRunner]{
				SuccessURL: lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website_runner.validate", Value: websiteRunnerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerUpdateView",
		lago.GetPageView("seer_websites.WebsiteRunnerUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.detail_update", views.LayerDetail[WebsiteRunner]{
				Key:          getters.Static("websiteRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_websites.website_runner.update", views.LayerUpdate[WebsiteRunner]{
				Key: getters.Static("websiteRunner"),
				SuccessURL: lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteRunner.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "seer_websites.website_runner.validate", Value: websiteRunnerValidate{}},
				},
			}))

	lago.RegistryView.Register("seer_websites.WebsiteRunnerDeleteView",
		lago.GetPageView("seer_websites.WebsiteRunnerDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("seer_websites.website_runner.delete_detail", views.LayerDetail[WebsiteRunner]{
				Key:          getters.Static("websiteRunner"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("seer_websites.website_runner.delete", views.LayerDelete[WebsiteRunner]{
				Key:        getters.Static("websiteRunner"),
				SuccessURL: lago.RoutePath("seer_websites.WebsiteRunnerListRoute", nil),
			}))
}

// websiteSourcePageURLFormPatcher turns POSTed URL strings into [lago.PageURL] so
// [views.PopulateFromMap] / mapstructure do not try to decode into embedded [url.URL].
type websiteSourcePageURLFormPatcher struct{}

func (websiteSourcePageURLFormPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if len(formErrors) > 0 {
		return formData, formErrors
	}
	raw, ok := formData["URL"]
	if !ok || raw == nil {
		return formData, formErrors
	}
	if _, ok := raw.(lago.PageURL); ok {
		return formData, formErrors
	}
	s, ok := raw.(string)
	if !ok {
		formErrors["URL"] = fmt.Errorf("invalid URL value type")
		return formData, formErrors
	}
	s = strings.TrimSpace(s)
	if s == "" {
		formData["URL"] = lago.PageURL{}
		return formData, formErrors
	}
	u, err := normalizeWebsiteURL(s)
	if err != nil {
		formErrors["URL"] = err
		return formData, formErrors
	}
	if urlFailsSSRF(r.Context(), u) {
		formErrors["URL"] = fmt.Errorf("url blocked by ssrf guard")
		return formData, formErrors
	}
	var pp lago.PageURL
	pp.SetFromURL(u)
	formData["URL"] = pp
	return formData, formErrors
}

type websiteSourceValidate struct{}

func (websiteSourceValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, _ := formData["URL"].(string)
	if strings.TrimSpace(raw) == "" {
		formErrors["URL"] = errors.New("seed URL is required")
	}
	d, ok := formData["Depth"].(uint)
	if !ok {
		if n64, ok2 := formData["Depth"].(uint64); ok2 {
			d = uint(n64)
		} else if n32, ok3 := formData["Depth"].(uint32); ok3 {
			d = uint(n32)
		} else if nint, ok4 := formData["Depth"].(int); ok4 && nint >= 0 {
			d = uint(nint)
		}
	}
	if d > maxWebsiteSourceDepth {
		formErrors["Depth"] = errors.New("depth is too large")
	}
	return formData, formErrors
}

type websiteRunnerValidate struct{}

func (websiteRunnerValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	name, _ := formData["Name"].(string)
	if strings.TrimSpace(name) == "" {
		formErrors["Name"] = errors.New("name is required")
	}
	durRaw, ok := formData["Duration"]
	if !ok {
		formErrors["Duration"] = errors.New("duration is required")
		return formData, formErrors
	}
	d, ok := durRaw.(*time.Duration)
	if !ok {
		formErrors["Duration"] = errors.New("invalid duration")
		return formData, formErrors
	}
	if d == nil || *d <= 0 {
		formErrors["Duration"] = errors.New("duration must be positive")
	}
	return formData, formErrors
}
