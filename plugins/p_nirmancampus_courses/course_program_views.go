package p_nirmancampus_courses

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func courseProgramQueryPatcherPreloads() []views.QueryPatcher {
	return []views.QueryPatcher{
		views.QueryPatcherPreload("Course"),
		views.QueryPatcherPreload("Program"),
		views.QueryPatcherOrderBy("semester ASC"),
	}
}

func parseCourseProgramDraft(r *http.Request) CourseProgram {
	parseUint := func(key string) uint {
		raw := r.URL.Query().Get(key)
		if raw == "" {
			return 0
		}
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return 0
		}
		return uint(value)
	}

	return CourseProgram{
		CourseID:  parseUint("CourseID"),
		ProgramID: parseUint("ProgramID"),
		Semester:  parseUint("Semester"),
	}
}

func withCourseProgramCreatePrefill(v *views.View) *views.View {
	return v.WithMethod(http.MethodGet, func(innerView *views.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			draft := parseCourseProgramDraft(r)
			ctx := r.Context()
			if draft.CourseID != 0 || draft.ProgramID != 0 || draft.Semester != 0 {
				ctx = context.WithValue(ctx, "courseprogram", draft)
			}
			innerView.RenderPage(w, r.WithContext(ctx))
		})
	})
}

func withCourseProgramQueryPatchers(v *views.View) *views.View {
	for i, patcher := range courseProgramQueryPatcherPreloads() {
		v = v.WithQueryPatcher("courses.courseprogram.preload."+strconv.Itoa(i), patcher)
	}
	return v
}

func init() {
	lago.RegistryView.Register("courses.CourseProgramListView",
		withCourseProgramQueryPatchers(
			views.ListView[CourseProgram]("courseprograms")(
				lago.GetPageView("courses.CourseProgramTable"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.CourseProgramDetailView",
		withCourseProgramQueryPatchers(
			views.DetailView[CourseProgram]("courseprogram")(
				lago.GetPageView("courses.CourseProgramDetail"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.CourseProgramCreateView",
		withCourseProgramCreatePrefill(
			views.CreateView[CourseProgram](
				lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("courses.CourseProgramCreateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.CourseProgramUpdateView",
		withCourseProgramQueryPatchers(
			views.DetailView[CourseProgram]("courseprogram")(
				views.UpdateView[CourseProgram](
					lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
					}),
				)(
					lago.GetPageView("courses.CourseProgramUpdateForm"),
				),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.CourseProgramDeleteView",
		withCourseProgramQueryPatchers(
			views.DetailView[CourseProgram]("courseprogram")(
				views.DeleteView[CourseProgram](
					lago.GetterRoutePath("courses.CourseProgramDefaultRoute", nil),
				)(
					lago.GetPageView("courses.CourseProgramDeleteForm"),
				),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.CourseProgramSelectView",
		withCourseProgramQueryPatchers(
			views.ListView[CourseProgram]("courseprograms")(
				lago.GetPageView("courses.CourseProgramSelectionTable"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
}
