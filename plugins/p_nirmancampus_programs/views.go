package p_nirmancampus_programs

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	baseviews "github.com/lariv-in/lago/views"
	"github.com/lariv-in/lago/p_programs"
	"gorm.io/gorm"
)

func fieldDBName[T any](db *gorm.DB, fieldName string) (string, bool) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", false
	}
	if stmt.Schema == nil {
		return "", false
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", false
	}
	return field.DBName, true
}

func regularValuesWithoutUniversity(values map[string]any) map[string]any {
	out := make(map[string]any, len(values))
	for k, v := range values {
		if k == "University" {
			continue
		}
		out[k] = v
	}
	return out
}

func universityFromValues(values map[string]any) string {
	raw := values["University"]
	s, _ := raw.(string)
	return strings.TrimSpace(s)
}

func upsertProgramUniversity(tx *gorm.DB, programID uint, values map[string]any) error {
	university := universityFromValues(values)

	var existing NirmancampusProgramDetails
	err := tx.Where("program_id = ?", programID).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&NirmancampusProgramDetails{
				ProgramID:  programID,
				University: university,
			}).Error
		}
		return err
	}

	existing.University = university
	return tx.Save(&existing).Error
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

func queryPatcherProgramUniversity(param string) baseviews.QueryPatcher {
	return func(_ *baseviews.View, r *http.Request, query *gorm.DB) *gorm.DB {
		getMap, ok := r.Context().Value("$get").(map[string]any)
		if !ok {
			return query
		}

		raw, ok := getMap[param]
		if !ok {
			return query
		}
		value, ok := raw.(string)
		if !ok {
			return query
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return query
		}

		universityDBName, ok := fieldDBName[NirmancampusProgramDetails](query, "University")
		if !ok {
			return query
		}
		programIDDBName, ok := fieldDBName[NirmancampusProgramDetails](query, "ProgramID")
		if !ok {
			return query
		}

		subquery := query.Session(&gorm.Session{NewDB: true}).
			Model(new(NirmancampusProgramDetails)).
			Select(programIDDBName).
			Where(universityDBName+" = ?", value)

		return query.Where("id IN (?)", subquery)
	}
}

func createProgramWithUniversity(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			regularValues := regularValuesWithoutUniversity(values)

			record := new(p_programs.Program)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, regularValues); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				return upsertProgramUniversity(tx, record.ID, values)
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

func updateProgramWithUniversity(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			regularValues := regularValuesWithoutUniversity(values)

			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(p_programs.Program)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}

				record := new(p_programs.Program)
				if err := query.First(record).Error; err != nil {
					return err
				}

				if len(regularValues) > 0 {
					updateQuery := tx.Model(new(p_programs.Program)).Where("id = ?", id)
					for _, queryPatcher := range view.QueryPatchers {
						updateQuery = queryPatcher.Value(view, r, updateQuery)
					}

					if err := updateQuery.Updates(regularValues).Error; err != nil {
						return err
					}
				}

				return upsertProgramUniversity(tx, uint(id), values)
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

func patchProgramViews() {
	successURL := lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	})

	lago.RegistryView.Patch("programs.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createProgramWithUniversity(successURL)
		return view
	})

	lago.RegistryView.Patch("programs.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateProgramWithUniversity(successURL)
		return view
	})

	lago.RegistryView.Patch("programs.ListView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher("nirmancampus_programs.filter_university", queryPatcherProgramUniversity("University"))
	})

	lago.RegistryView.Patch("programs.SelectView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher("nirmancampus_programs.filter_university", queryPatcherProgramUniversity("University"))
	})
}

func init() {
	patchProgramViews()
}
