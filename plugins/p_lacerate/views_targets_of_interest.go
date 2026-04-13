package p_lacerate

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const ctxKeyRelatedTargetsOfInterest = "relatedTargetsOfInterest"
const ctxKeyRelatedReports = "relatedReports"
const ctxKeyRelatedIntel = "relatedIntel"

func loadIntelListWithSource(ctx context.Context, db *gorm.DB, rows []Intel) ([]Intel, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.ID != 0 {
			ids = append(ids, row.ID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var loaded []Intel
	if err := db.WithContext(ctx).Preload("Source").Where("id IN ?", ids).Find(&loaded).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]Intel, len(loaded))
	for _, row := range loaded {
		byID[row.ID] = row
	}
	out := make([]Intel, 0, len(rows))
	for _, row := range rows {
		if loadedRow, ok := byID[row.ID]; ok {
			out = append(out, loadedRow)
		}
	}
	return out, nil
}

type targetOfInterestRelatedLayer struct{}

func (targetOfInterestRelatedLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		target, ok := ctx.Value("target_of_interest").(TargetOfInterest)
		if !ok || target.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			slog.Error("lacerate: related targets of interest: missing db in context")
			next.ServeHTTP(w, r)
			return
		}
		if target.Embedding == nil {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rows, err := searchTargetsOfInterestByEmbedding(db.WithContext(ctx), *target.Embedding, 7)
		if err != nil {
			slog.Error("lacerate: related targets of interest search", "error", err, "target_of_interest_id", target.ID)
			next.ServeHTTP(w, r)
			return
		}
		related := make([]TargetOfInterest, 0, 6)
		for _, row := range rows {
			if row.ID == 0 || row.ID == target.ID {
				continue
			}
			related = append(related, row)
			if len(related) == 6 {
				break
			}
		}
		ctx = context.WithValue(ctx, ctxKeyRelatedTargetsOfInterest, components.ObjectList[TargetOfInterest]{
			Items:    related,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(related)),
		})
		reportRows, err := searchReportsByEmbedding(db.WithContext(ctx), *target.Embedding, 6)
		if err != nil {
			slog.Error("lacerate: related reports search", "error", err, "target_of_interest_id", target.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		reportItems, err := loadReportPageDataList(ctx, db, reportRows)
		if err != nil {
			slog.Error("lacerate: related reports page data", "error", err, "target_of_interest_id", target.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, ctxKeyRelatedReports, components.ObjectList[ReportPageData]{
			Items:    reportItems,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(reportItems)),
		})
		intelRows, err := searchIntelByEmbedding(db.WithContext(ctx), *target.Embedding, 6)
		if err != nil {
			slog.Error("lacerate: related intel search", "error", err, "target_of_interest_id", target.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		intelItems, err := loadIntelListWithSource(ctx, db, intelRows)
		if err != nil {
			slog.Error("lacerate: related intel load", "error", err, "target_of_interest_id", target.ID)
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
	patchers := views.QueryPatchers[TargetOfInterest]{
		{Key: "lacerate.targets_of_interest.order_id", Value: views.QueryPatcherOrderBy[TargetOfInterest]{Order: "id DESC"}},
	}

	lago.RegistryView.Register("lacerate.TargetOfInterestListView",
		lago.GetPageView("lacerate.TargetsOfInterestTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.list", views.LayerList[TargetOfInterest]{
				Key:           getters.Static("targets_of_interest"),
				QueryPatchers: patchers,
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestDetailView",
		lago.GetPageView("lacerate.TargetOfInterestDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.targets_of_interest.related", targetOfInterestRelatedLayer{}))

	lago.RegistryView.Register("lacerate.TargetOfInterestCreateView",
		lago.GetPageView("lacerate.TargetOfInterestCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.create", views.LayerCreate[TargetOfInterest]{
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestUpdateView",
		lago.GetPageView("lacerate.TargetOfInterestUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.update_detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.targets_of_interest.update", views.LayerUpdate[TargetOfInterest]{
				Key: getters.Static("target_of_interest"),
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
				}),
			}))

	lago.RegistryView.Register("lacerate.TargetOfInterestDeleteView",
		lago.GetPageView("lacerate.TargetOfInterestDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("lacerate.targets_of_interest.delete_detail", views.LayerDetail[TargetOfInterest]{
				Key:          getters.Static("target_of_interest"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("lacerate.targets_of_interest.delete", views.LayerDelete[TargetOfInterest]{
				Key:        getters.Static("target_of_interest"),
				SuccessURL: lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
			}))
}
