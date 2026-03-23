package p_academicrecords_courses

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_academicrecords"
	"github.com/lariv-in/lago/p_academicrecords_programs"
	baseviews "github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func joinFieldDBName[T any](db *gorm.DB, fieldName string) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", err
	}
	if stmt.Schema == nil {
		return "", fmt.Errorf("schema not found for %T", new(T))
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", fmt.Errorf("field %q not found for %T", fieldName, new(T))
	}
	return field.DBName, nil
}

func associationIDsFromValues(values map[string]any, field string) []uint {
	raw, ok := values[field]
	if !ok {
		return nil
	}
	switch typed := raw.(type) {
	case components.AssociationIDs:
		return typed.IDs
	case *components.AssociationIDs:
		if typed == nil {
			return nil
		}
		return typed.IDs
	case []uint:
		return typed
	default:
		return nil
	}
}

func regularValuesWithout(values map[string]any, field string) map[string]any {
	regular := make(map[string]any, len(values))
	for key, value := range values {
		if key == field {
			continue
		}
		regular[key] = value
	}
	return regular
}

func syncJoinRows[TJoin any](tx *gorm.DB, ownerField, relatedField string, ownerID uint, relatedIDs []uint) error {
	ownerDBName, err := joinFieldDBName[TJoin](tx, ownerField)
	if err != nil {
		return err
	}

	if err := tx.Where(ownerDBName+" = ?", ownerID).Delete(new(TJoin)).Error; err != nil {
		return err
	}
	if len(relatedIDs) == 0 {
		return nil
	}

	deduped := make([]uint, 0, len(relatedIDs))
	seen := map[uint]struct{}{}
	for _, id := range relatedIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	if len(deduped) == 0 {
		return nil
	}

	rows := make([]TJoin, 0, len(deduped))
	for _, relatedID := range deduped {
		row := new(TJoin)
		if err := baseviews.PopulateFromMap(row, map[string]any{
			ownerField:   ownerID,
			relatedField: relatedID,
		}); err != nil {
			return err
		}
		rows = append(rows, *row)
	}
	return tx.Create(&rows).Error
}

func loadRecordForContext[T any](view *baseviews.View, r *http.Request) (*T, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		return nil, err
	}
	db := r.Context().Value("$db").(*gorm.DB)
	query := db.Model(new(T)).Where("id = ?", id)
	for _, queryPatcher := range view.QueryPatchers {
		query = queryPatcher.Value(view, r, query)
	}
	record := new(T)
	if err := query.First(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func renderUpdateErrorsWithRecord[T any](w http.ResponseWriter, r *http.Request, view *baseviews.View, contextKey string, fieldErrors map[string]error, values map[string]any) {
	record, err := loadRecordForContext[T](view, r)
	if err != nil {
		view.RenderWithErrors(w, r, fieldErrors, values)
		return
	}
	ctx := context.WithValue(r.Context(), contextKey, *record)
	view.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
}

func createAcademicRecordWithProgramAndCourses(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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

			programID := p_academicrecords_programs.ProgramIDFromValues(values)
			relatedIDs := associationIDsFromValues(values, coursesFieldName)
			baseValues := p_academicrecords_programs.RegularValuesWithoutProgram(regularValuesWithout(values, coursesFieldName))

			record := new(p_academicrecords.AcademicRecord)
			db := r.Context().Value("$db").(*gorm.DB)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, baseValues); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				if err := p_academicrecords_programs.UpsertAcademicRecordProgram(tx, record.ID, programID); err != nil {
					return err
				}
				return syncJoinRows[AcademicRecordCourse](tx, "AcademicRecordID", "CourseID", record.ID, relatedIDs)
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

func updateAcademicRecordWithProgramAndCourses(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
	return func(view *baseviews.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values, fieldErrors, err := view.ParseForm(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if view.HasErrors(fieldErrors) {
				renderUpdateErrorsWithRecord[p_academicrecords.AcademicRecord](w, r, view, "academicrecord", fieldErrors, values)
				return
			}

			id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			programID := p_academicrecords_programs.ProgramIDFromValues(values)
			relatedIDs := associationIDsFromValues(values, coursesFieldName)
			baseValues := p_academicrecords_programs.RegularValuesWithoutProgram(regularValuesWithout(values, coursesFieldName))

			db := r.Context().Value("$db").(*gorm.DB)
			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(p_academicrecords.AcademicRecord)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}
				record := new(p_academicrecords.AcademicRecord)
				if err := query.First(record).Error; err != nil {
					return err
				}
				if len(baseValues) > 0 {
					updateQuery := tx.Model(new(p_academicrecords.AcademicRecord)).Where("id = ?", id)
					for _, queryPatcher := range view.QueryPatchers {
						updateQuery = queryPatcher.Value(view, r, updateQuery)
					}
					if err := updateQuery.Updates(baseValues).Error; err != nil {
						return err
					}
				}
				if err := p_academicrecords_programs.UpsertAcademicRecordProgram(tx, uint(id), programID); err != nil {
					return err
				}
				return syncJoinRows[AcademicRecordCourse](tx, "AcademicRecordID", "CourseID", uint(id), relatedIDs)
			})
			if err != nil {
				fieldErrors["_form"] = err
				renderUpdateErrorsWithRecord[p_academicrecords.AcademicRecord](w, r, view, "academicrecord", fieldErrors, values)
				return
			}

			ctx := context.WithValue(r.Context(), "$id", uint(id))
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func patchAcademicRecordViews() {
	successURL := lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	})

	lago.RegistryView.Patch("academicrecords.ListView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher(
			"academicrecords_courses.filter_courses",
			baseviews.QueryPatcherJoinFilter[AcademicRecordCourse](coursesFieldName, "AcademicRecordID", "CourseID"),
		)
	})

	lago.RegistryView.Patch("academicrecords.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createAcademicRecordWithProgramAndCourses(successURL)
		return view
	})

	lago.RegistryView.Patch("academicrecords.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateAcademicRecordWithProgramAndCourses(successURL)
		return view
	})
}

func init() {
	patchAcademicRecordViews()
}
