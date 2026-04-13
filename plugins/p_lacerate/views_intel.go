package p_lacerate

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const ctxKeyIntelEvents = "intelEvents"

func ctxWithIntelEvents(ctx context.Context, db *gorm.DB, intelID uint) context.Context {
	if db == nil || db.Name() != "postgres" {
		return context.WithValue(ctx, ctxKeyIntelEvents, components.ObjectList[Event]{
			Number:   1,
			NumPages: 1,
		})
	}
	var evs []Event
	if err := db.WithContext(ctx).Where("intel_id = ?", intelID).Order("datetime DESC, id DESC").Find(&evs).Error; err != nil {
		slog.Error("lacerate: load intel events", "error", err, "intel_id", intelID)
		return context.WithValue(ctx, ctxKeyIntelEvents, components.ObjectList[Event]{
			Number:   1,
			NumPages: 1,
		})
	}
	return context.WithValue(ctx, ctxKeyIntelEvents, components.ObjectList[Event]{
		Items:    evs,
		Number:   1,
		NumPages: 1,
		Total:    uint64(len(evs)),
	})
}

type intelRelatedLayer struct{}

func (intelRelatedLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		intel, ok := ctx.Value("intel").(Intel)
		if !ok || intel.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db := ctx.Value("$db").(*gorm.DB)
		ctx = ctxWithIntelEvents(ctx, db, intel.ID)
		targetRows, err := searchTargetsOfInterestByEmbedding(db.WithContext(ctx), intel.Embedding, 6)
		if err != nil {
			slog.Error("lacerate: intel related targets search", "error", err, "intel_id", intel.ID)
			next.ServeHTTP(w, r)
			return
		}
		ctx = context.WithValue(ctx, ctxKeyRelatedTargetsOfInterest, components.ObjectList[TargetOfInterest]{
			Items:    targetRows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(targetRows)),
		})
		reportRows, err := searchReportsByEmbedding(db.WithContext(ctx), intel.Embedding, 6)
		if err != nil {
			slog.Error("lacerate: intel related reports search", "error", err, "intel_id", intel.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		reportItems, err := loadReportPageDataList(ctx, db, reportRows)
		if err != nil {
			slog.Error("lacerate: intel related reports page data", "error", err, "intel_id", intel.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, ctxKeyRelatedReports, components.ObjectList[ReportPageData]{
			Items:    reportItems,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(reportItems)),
		})
		intelRows, err := searchIntelByEmbedding(db.WithContext(ctx), intel.Embedding, 7)
		if err != nil {
			slog.Error("lacerate: intel related intel search", "error", err, "intel_id", intel.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		relatedIntel := make([]Intel, 0, 6)
		for _, row := range intelRows {
			if row.ID == 0 || row.ID == intel.ID {
				continue
			}
			relatedIntel = append(relatedIntel, row)
			if len(relatedIntel) == 6 {
				break
			}
		}
		intelItems, err := loadIntelListWithSource(ctx, db, relatedIntel)
		if err != nil {
			slog.Error("lacerate: intel related intel load", "error", err, "intel_id", intel.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, ctxKeyRelatedIntel, components.ObjectList[Intel]{
			Items:    intelItems,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(intelItems)),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	intelListPatchers := views.QueryPatchers[Intel]{
		{Key: "lacerate.intel.preload_source", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
		{Key: "lacerate.intel.preload_preview", Value: views.QueryPatcherPreload[Intel]{Field: "PreviewImage"}},
		{Key: "lacerate.intel.order_id", Value: views.QueryPatcherOrderBy[Intel]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.IntelListView",
		lago.GetPageView("lacerate.IntelTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.list", views.LayerList[Intel]{
				Key:           getters.Static("intels"),
				QueryPatchers: intelListPatchers,
			}))

	lago.RegistryView.Register("lacerate.IntelDetailView",
		lago.GetPageView("lacerate.IntelDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.detail_preload", Value: views.QueryPatcherPreload[Intel]{
						Field: "Source",
					}},
					{Key: "lacerate.intel.detail_preload_preview", Value: views.QueryPatcherPreload[Intel]{
						Field: "PreviewImage",
					}},
				},
			}).
			WithLayer("lacerate.intel.related", intelRelatedLayer{}))

	lago.RegistryView.Register("lacerate.IntelCreateView",
		lago.GetPageView("lacerate.IntelCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.create", views.LayerCreate[Intel]{
				SuccessURL: lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.IntelUpdateView",
		lago.GetPageView("lacerate.IntelUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.update_detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.update_preload_src", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
					{Key: "lacerate.intel.update_preload_preview", Value: views.QueryPatcherPreload[Intel]{Field: "PreviewImage"}},
				},
			}).
			WithLayer("lacerate.intel.update", views.LayerUpdate[Intel]{
				Key: getters.Static("intel"),
				SuccessURL: lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.IntelDeleteView",
		lago.GetPageView("lacerate.IntelDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.intel.delete_detail", views.LayerDetail[Intel]{
				Key:          getters.Static("intel"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Intel]{
					{Key: "lacerate.intel.delete_preload", Value: views.QueryPatcherPreload[Intel]{Field: "Source"}},
				},
			}).
			WithLayer("lacerate.intel.delete", views.LayerDelete[Intel]{
				Key:        getters.Static("intel"),
				SuccessURL: lago.RoutePath("lacerate.IntelListRoute", nil),
			}))

	lago.RegistryView.Register("lacerate.SourceSelectView",
		lago.GetPageView("lacerate.SourceSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.sources.select", views.LayerList[Source]{
				Key: getters.Static("sources"),
				QueryPatchers: views.QueryPatchers[Source]{
					registry.Pair[string, views.QueryPatcher[Source]]{
						Key:   "lacerate.sources.order_id",
						Value: views.QueryPatcherOrderBy[Source]{Order: "id DESC"},
					},
				},
			}))
}
