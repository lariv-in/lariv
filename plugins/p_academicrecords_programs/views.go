package p_academicrecords_programs

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	baseviews "github.com/lariv-in/views"
	"github.com/lariv-in/p_academicrecords"
	"gorm.io/gorm"
)

func programIDFromValues(values map[string]any) uint {
	raw, ok := values["ProgramID"]
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

func splitRegularValues(values map[string]any) map[string]any {
	regularValues := make(map[string]any, len(values))
	for k, v := range values {
		if k == "ProgramID" {
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

func createAcademicRecordWithProgram(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			programID := programIDFromValues(values)
			regularValues := splitRegularValues(values)

			record := new(p_academicrecords.AcademicRecord)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, regularValues); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				return upsertAcademicRecordProgram(tx, record.ID, programID)
			})

			if err != nil {
				fieldErrors["_form"] = err
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			ctx := context.WithValue(r.Context(), "$id", record.ID)
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func updateAcademicRecordWithProgram(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			programID := programIDFromValues(values)
			regularValues := splitRegularValues(values)

			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(p_academicrecords.AcademicRecord)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}

				record := new(p_academicrecords.AcademicRecord)
				if err := query.First(record).Error; err != nil {
					return err
				}

				if len(regularValues) > 0 {
					updateQuery := tx.Model(new(p_academicrecords.AcademicRecord)).Where("id = ?", id)
					for _, queryPatcher := range view.QueryPatchers {
						updateQuery = queryPatcher.Value(view, r, updateQuery)
					}

					if err := updateQuery.Updates(regularValues).Error; err != nil {
						return err
					}
				}

				return upsertAcademicRecordProgram(tx, uint(id), programID)
			})
			if err != nil {
				fieldErrors["_form"] = err
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			ctx := context.WithValue(r.Context(), "$id", uint(id))
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func patchAcademicRecordViews() {
	const successDetail = "academicrecords.DetailRoute"
	successURL := lago.GetterRoutePath(successDetail, map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	})

	lago.RegistryView.Patch("academicrecords.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createAcademicRecordWithProgram(successURL)
		return view
	})

	lago.RegistryView.Patch("academicrecords.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateAcademicRecordWithProgram(successURL)
		return view
	})
}

func init() {
	patchAcademicRecordViews()
}

