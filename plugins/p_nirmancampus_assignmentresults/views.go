package p_nirmancampus_assignmentresults

import (
	"context"
	"log/slog"
	"maps"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func assignmentResultsOrderIDDesc(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	if r.URL.Query().Get("sort") != "" {
		return query
	}
	return query.Order("id DESC")
}

// assignmentDetailLoadResultsMiddleware adds a paginated ObjectList for this assignment's results
// before the assignments DetailView handler runs (same context key as global list views).
func assignmentDetailLoadResultsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		idStr := r.PathValue("id")
		if idStr == "" {
			next.ServeHTTP(w, r)
			return
		}
		aid, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		db := r.Context().Value("$db").(*gorm.DB)
		const pageSize = 12
		pageNum := 1
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				pageNum = p
			}
		}
		countQuery := db.Model(&AssignmentResult{}).Where("assignment_id = ?", aid)
		var total int64
		if err := countQuery.Count(&total).Error; err != nil {
			slog.Error("assignmentresults: count on assignment detail failed", "err", err)
			next.ServeHTTP(w, r)
			return
		}
		q := db.Model(&AssignmentResult{}).Where("assignment_id = ?", aid).
			Preload("AcademicRecord").Preload("AcademicRecord.Student.User")
		q = assignmentResultsOrderIDDesc(nil, r, q)
		q = q.Limit(pageSize).Offset((pageNum - 1) * pageSize)
		var results []AssignmentResult
		if err := q.Find(&results).Error; err != nil {
			slog.Error("assignmentresults: list on assignment detail failed", "err", err)
			next.ServeHTTP(w, r)
			return
		}
		numPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		if numPages == 0 {
			numPages = 1
		}
		objectList := components.ObjectList[AssignmentResult]{
			Items:    results,
			Number:   pageNum,
			NumPages: numPages,
			Total:    total,
		}
		ctx := context.WithValue(r.Context(), "assignmentresults", objectList)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createFormPrefillAssignmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		aid := r.URL.Query().Get("AssignmentID")
		if aid == "" {
			next.ServeHTTP(w, r)
			return
		}
		id, err := strconv.ParseUint(aid, 10, 64)
		if err != nil {
			slog.Error("assignmentresults: invalid AssignmentID in query", "AssignmentID", aid, "err", err)
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		inMap := map[string]any{}
		if existing, ok := ctx.Value(getters.ContextKeyIn).(map[string]any); ok {
			maps.Copy(inMap, existing)
		}
		inMap["AssignmentID"] = uint(id)
		ctx = context.WithValue(ctx, getters.ContextKeyIn, inMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	lago.RegistryView.Register("assignmentresults.ListView",
		views.ListView[AssignmentResult]("assignmentresults")(
			lago.GetPageView("assignmentresults.AssignmentResultTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentresults.preload_assignment", views.QueryPatcherPreload("Assignment")).
			WithQueryPatcher("assignmentresults.preload_academic_record", views.QueryPatcherPreload("AcademicRecord")).
			WithQueryPatcher("assignmentresults.preload_academic_record_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")).
			WithQueryPatcher("assignmentresults.order_id_desc", assignmentResultsOrderIDDesc))

	lago.RegistryView.Register("assignmentresults.DetailView",
		views.DetailView[AssignmentResult]("assignmentresult")(
			lago.GetPageView("assignmentresults.AssignmentResultDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentresults.preload_assignment", views.QueryPatcherPreload("Assignment")).
			WithQueryPatcher("assignmentresults.preload_academic_record", views.QueryPatcherPreload("AcademicRecord")).
			WithQueryPatcher("assignmentresults.preload_academic_record_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")))

	lago.RegistryView.Register("assignmentresults.CreateView",
		views.CreateView[AssignmentResult](
			lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("assignmentresults.AssignmentResultCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("assignmentresults.UpdateView",
		views.DetailView[AssignmentResult]("assignmentresult")(
			views.UpdateView[AssignmentResult](
				lago.GetterRoutePath("assignmentresults.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("assignmentresults.AssignmentResultUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentresults.preload_assignment", views.QueryPatcherPreload("Assignment")).
			WithQueryPatcher("assignmentresults.preload_academic_record", views.QueryPatcherPreload("AcademicRecord")).
			WithQueryPatcher("assignmentresults.preload_academic_record_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")))

	lago.RegistryView.Register("assignmentresults.DeleteView",
		views.DetailView[AssignmentResult]("assignmentresult")(
			views.DeleteView[AssignmentResult](
				lago.GetterRoutePath("assignmentresults.DefaultRoute", nil),
			)(
				lago.GetPageView("assignmentresults.AssignmentResultDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentresults.preload_assignment", views.QueryPatcherPreload("Assignment")).
			WithQueryPatcher("assignmentresults.preload_academic_record", views.QueryPatcherPreload("AcademicRecord")).
			WithQueryPatcher("assignmentresults.preload_academic_record_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")))

	lago.RegistryView.Register("assignmentresults.SelectView",
		views.ListView[AssignmentResult]("assignmentresults")(
			lago.GetPageView("assignmentresults.AssignmentResultSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignmentresults.preload_assignment", views.QueryPatcherPreload("Assignment")).
			WithQueryPatcher("assignmentresults.preload_academic_record", views.QueryPatcherPreload("AcademicRecord")).
			WithQueryPatcher("assignmentresults.preload_academic_record_student_user", views.QueryPatcherPreload("AcademicRecord.Student.User")).
			WithQueryPatcher("assignmentresults.order_id_desc", assignmentResultsOrderIDDesc))

	lago.RegistryView.Patch("assignments.DetailView", func(view *views.View) *views.View {
		return view.WithMiddleware("assignmentresults.detail_load_results", assignmentDetailLoadResultsMiddleware)
	})

	lago.RegistryView.Patch("assignmentresults.CreateView", func(view *views.View) *views.View {
		return view.WithMiddleware("assignmentresults.prefill_create_assignment", createFormPrefillAssignmentMiddleware)
	})
}
