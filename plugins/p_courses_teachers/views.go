package p_courses_teachers

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_courses"
	"github.com/lariv-in/p_teachers"
	baseviews "github.com/lariv-in/views"
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

func modelID[T any](record *T) (uint, error) {
	value := reflect.ValueOf(record)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return 0, fmt.Errorf("record is nil")
	}
	idField := value.Elem().FieldByName("ID")
	if !idField.IsValid() {
		return 0, fmt.Errorf("record %T does not have an ID field", record)
	}
	switch idField.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint(idField.Uint()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint(idField.Int()), nil
	default:
		return 0, fmt.Errorf("record %T has unsupported ID field kind %s", record, idField.Kind())
	}
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

func createWithJoinHandler[T any, TJoin any](field, ownerField, relatedField string, successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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

			relatedIDs := associationIDsFromValues(values, field)
			record := new(T)
			db := r.Context().Value("$db").(*gorm.DB)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, regularValuesWithout(values, field)); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				id, err := modelID(record)
				if err != nil {
					return err
				}
				return syncJoinRows[TJoin](tx, ownerField, relatedField, id, relatedIDs)
			})
			if err != nil {
				fieldErrors["_form"] = err
				view.RenderWithErrors(w, r, fieldErrors, values)
				return
			}

			id, err := modelID(record)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), "$id", id)
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func updateWithJoinHandler[T any, TJoin any](contextKey, field, ownerField, relatedField string, successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
	return func(view *baseviews.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values, fieldErrors, err := view.ParseForm(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if view.HasErrors(fieldErrors) {
				renderUpdateErrorsWithRecord[T](w, r, view, contextKey, fieldErrors, values)
				return
			}

			id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			relatedIDs := associationIDsFromValues(values, field)
			db := r.Context().Value("$db").(*gorm.DB)
			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(T)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}
				record := new(T)
				if err := query.First(record).Error; err != nil {
					return err
				}
				regularValues := regularValuesWithout(values, field)
				if len(regularValues) > 0 {
					if err := tx.Model(new(T)).Where("id = ?", id).Updates(regularValues).Error; err != nil {
						return err
					}
				}
				return syncJoinRows[TJoin](tx, ownerField, relatedField, uint(id), relatedIDs)
			})
			if err != nil {
				fieldErrors["_form"] = err
				renderUpdateErrorsWithRecord[T](w, r, view, contextKey, fieldErrors, values)
				return
			}

			ctx := context.WithValue(r.Context(), "$id", uint(id))
			redirectURL, _ := getters.IfOrGetter(successURL, ctx, "")
			lago.Redirect(w, r, redirectURL)
		})
	}
}

func patchCourseViews() {
	lago.RegistryView.Patch("courses.ListView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher(
			"courses_teachers.filter_teachers",
			baseviews.QueryPatcherJoinFilter[CourseTeacher](teachersFieldName, "CourseID", "TeacherID"),
		)
	})

	lago.RegistryView.Patch("courses.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createWithJoinHandler[p_courses.Course, CourseTeacher](
			teachersFieldName,
			"CourseID",
			"TeacherID",
			lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)
		return view
	})

	lago.RegistryView.Patch("courses.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateWithJoinHandler[p_courses.Course, CourseTeacher](
			"course",
			teachersFieldName,
			"CourseID",
			"TeacherID",
			lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)
		return view
	})
}

func patchTeacherViews() {
	lago.RegistryView.Patch("teachers.ListView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher(
			"courses_teachers.filter_courses",
			baseviews.QueryPatcherJoinFilter[CourseTeacher](coursesFieldName, "TeacherID", "CourseID"),
		)
	})

	lago.RegistryView.Patch("teachers.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createWithJoinHandler[p_teachers.Teacher, CourseTeacher](
			coursesFieldName,
			"TeacherID",
			"CourseID",
			lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)
		return view
	})

	lago.RegistryView.Patch("teachers.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateWithJoinHandler[p_teachers.Teacher, CourseTeacher](
			"teacher",
			coursesFieldName,
			"TeacherID",
			"CourseID",
			lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)
		return view
	})
}

func init() {
	patchCourseViews()
	patchTeacherViews()
}
