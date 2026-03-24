package p_assignments_semesters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_assignments"
	"github.com/lariv-in/lago/p_semesters"
	baseviews "github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func parseSemesterEnvID(raw string) (uint, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	if i := strings.IndexByte(raw, ':'); i > 0 {
		raw = raw[:i]
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func semesterEnvironmentDefault(db *gorm.DB, now time.Time) (uint, bool) {
	var sem p_semesters.Semester
	err := db.Model(&p_semesters.Semester{}).
		Where(`"start" <= ? AND "end" >= ?`, now, now).
		Order(`"start" ASC`).
		First(&sem).Error
	if err != nil {
		return 0, false
	}
	return sem.ID, true
}

// assignmentsListSemesterEnvQueryPatcher scopes the list view to the semester
// selected in the environment cookie (components.Environment[uint] semester key).
func assignmentsListSemesterEnvQueryPatcher(_ *baseviews.View, r *http.Request, query *gorm.DB) *gorm.DB {
	envMap, ok := r.Context().Value("$environment").(map[string]string)
	if !ok {
		return query
	}
	raw, ok := envMap["semester"]
	if !ok || strings.TrimSpace(raw) == "" {
		db, dbOK := r.Context().Value("$db").(*gorm.DB)
		if !dbOK || db == nil {
			return query
		}
		id, found := semesterEnvironmentDefault(db, time.Now())
		if !found {
			return query
		}
		raw = fmt.Sprintf("%d", id)
	}
	semesterID, ok := parseSemesterEnvID(raw)
	if !ok {
		return query
	}

	sub := query.Session(&gorm.Session{NewDB: true}).Model(&AssignmentSemesterDetails{}).
		Select("assignment_id").
		Where("semester_id = ?", semesterID)

	return query.Where("id IN (?)", sub)
}

func SemesterIDFromValues(values map[string]any) uint {
	raw, ok := values["SemesterID"]
	if !ok || raw == nil {
		return 0
	}

	switch typed := raw.(type) {
	case uint:
		return typed
	case *uint:
		if typed == nil {
			return 0
		}
		return *typed
	case int:
		if typed <= 0 {
			return 0
		}
		return uint(typed)
	default:
		return 0
	}
}

func regularValuesWithoutSemesterID(values map[string]any) map[string]any {
	regularValues := make(map[string]any, len(values))
	for k, v := range values {
		if k == "SemesterID" {
			continue
		}
		regularValues[k] = v
	}
	return regularValues
}

func parseUintPathID(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	if raw == "" {
		return 0, errors.New("empty id")
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func createAssignmentWithSemester(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
	return func(view *baseviews.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values, fieldErrors, err := view.ParseForm(w, r)
			if err != nil {
				view.RenderWithErrors(w, r, map[string]error{"_form": err}, values)
				return
			}
			if view.HasErrors(fieldErrors) {
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			db := r.Context().Value("$db").(*gorm.DB)
			semesterID := SemesterIDFromValues(values)
			regularValues := regularValuesWithoutSemesterID(values)

			record := new(p_assignments.Assignment)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, regularValues); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				return UpsertAssignmentSemester(tx, record.ID, semesterID)
			})
			if err != nil {
				fieldErrors["_form"] = err
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			id := uint(reflect.ValueOf(*record).FieldByName("ID").Uint())
			ctx := context.WithValue(r.Context(), "$id", id)
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func updateAssignmentWithSemester(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
	return func(view *baseviews.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values, fieldErrors, err := view.ParseForm(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if view.HasErrors(fieldErrors) {
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			id, err := parseUintPathID(r, "id")
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			db := r.Context().Value("$db").(*gorm.DB)
			semesterID := SemesterIDFromValues(values)
			regularValues := regularValuesWithoutSemesterID(values)

			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(p_assignments.Assignment)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}

				record := new(p_assignments.Assignment)
				if err := query.First(record).Error; err != nil {
					return err
				}

				if len(regularValues) > 0 {
					updateQuery := tx.Model(new(p_assignments.Assignment)).Where("id = ?", id)
					for _, queryPatcher := range view.QueryPatchers {
						updateQuery = queryPatcher.Value(view, r, updateQuery)
					}

					if err := updateQuery.Updates(regularValues).Error; err != nil {
						return err
					}
				}

				return UpsertAssignmentSemester(tx, id, semesterID)
			})
			if err != nil {
				fieldErrors["_form"] = err
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			ctx := context.WithValue(r.Context(), "$id", id)
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func patchAssignmentViews() {
	successURL := lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	})

	lago.RegistryView.Patch("assignments.ListView", func(v *baseviews.View) *baseviews.View {
		return v.WithQueryPatcher("assignments_semesters.filter_env_semester", assignmentsListSemesterEnvQueryPatcher)
	})

	lago.RegistryView.Patch("assignments.SelectView", func(v *baseviews.View) *baseviews.View {
		return v.WithQueryPatcher("assignments_semesters.filter_env_semester", assignmentsListSemesterEnvQueryPatcher)
	})

	lago.RegistryView.Patch("assignments.CreateView", func(v *baseviews.View) *baseviews.View {
		v.Handlers[http.MethodPost] = createAssignmentWithSemester(successURL)
		return v
	})

	lago.RegistryView.Patch("assignments.UpdateView", func(v *baseviews.View) *baseviews.View {
		v.Handlers[http.MethodPost] = updateAssignmentWithSemester(successURL)
		return v
	})
}
