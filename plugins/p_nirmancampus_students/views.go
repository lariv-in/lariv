package p_nirmancampus_students

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	baseviews "github.com/lariv-in/lago/views"
	"github.com/lariv-in/lago/p_students"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func splitAssociationValues(values map[string]any) (map[string]any, map[string]components.AssociationIDs) {
	regularValues := make(map[string]any, len(values))
	associationValues := map[string]components.AssociationIDs{}
	for key, value := range values {
		switch typed := value.(type) {
		case components.AssociationIDs:
			if typed.Field == "" {
				typed.Field = key
			}
			associationValues[key] = typed
		case *components.AssociationIDs:
			if typed == nil {
				continue
			}
			if typed.Field == "" {
				typed.Field = key
			}
			associationValues[key] = *typed
		default:
			regularValues[key] = value
		}
	}
	return regularValues, associationValues
}

func applyAssociationReplacements(db *gorm.DB, record any, associations map[string]components.AssociationIDs) error {
	if len(associations) == 0 {
		return nil
	}
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(record); err != nil {
		return err
	}

	for _, associationValue := range associations {
		relationship, ok := stmt.Schema.Relationships.Relations[associationValue.Field]
		if !ok {
			return fmt.Errorf("unknown association field %q", associationValue.Field)
		}
		if relationship.Type != schema.Many2Many {
			return fmt.Errorf("field %q is not a many-to-many association", associationValue.Field)
		}

		association := db.Model(record).Association(associationValue.Field)
		if association.Error != nil {
			return association.Error
		}

		if len(associationValue.IDs) == 0 {
			if err := association.Clear(); err != nil {
				return err
			}
			continue
		}

		replaceValue, err := buildAssociationReplaceValue(relationship, associationValue.IDs)
		if err != nil {
			return err
		}
		if err := association.Replace(replaceValue); err != nil {
			return err
		}
	}

	return nil
}

func buildAssociationReplaceValue(relationship *schema.Relationship, ids []uint) (any, error) {
	sliceType := relationship.Field.FieldType
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field %q is not a slice association", relationship.Field.Name)
	}

	elemType := sliceType.Elem()
	elemIsPointer := elemType.Kind() == reflect.Pointer
	baseType := elemType
	if elemIsPointer {
		baseType = elemType.Elem()
	}

	sliceValue := reflect.MakeSlice(sliceType, 0, len(ids))
	for _, id := range ids {
		itemPtr := reflect.New(baseType)
		idField := itemPtr.Elem().FieldByName("ID")
		if !idField.IsValid() || !idField.CanSet() {
			return nil, fmt.Errorf("association %q element type %s does not have a settable ID field", relationship.Field.Name, baseType)
		}
		switch idField.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			idField.SetUint(uint64(id))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			idField.SetInt(int64(id))
		default:
			return nil, fmt.Errorf("association %q element ID field has unsupported kind %s", relationship.Field.Name, idField.Kind())
		}

		if elemIsPointer {
			sliceValue = reflect.Append(sliceValue, itemPtr)
		} else {
			sliceValue = reflect.Append(sliceValue, itemPtr.Elem())
		}
	}
	return sliceValue.Interface(), nil
}

func upsertStudentDetails(tx *gorm.DB, studentID uint, values map[string]any) error {
	fathersName, _ := values["FathersName"].(string)
	category, _ := values["Category"].(string)
	address, _ := values["Address"].(string)

	// Trim whitespace for nicer UX and to match typical form semantics.
	fathersName = strings.TrimSpace(fathersName)
	category = strings.TrimSpace(category)
	address = strings.TrimSpace(address)

	var existing NirmancampusStudentDetails
	err := tx.Where("student_id = ?", studentID).Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&NirmancampusStudentDetails{
				StudentID:    studentID,
				FathersName: fathersName,
				Category:    category,
				Address:     address,
			}).Error
		}
		return err
	}

	existing.FathersName = fathersName
	existing.Category = category
	existing.Address = address
	return tx.Save(&existing).Error
}

func removeExtensionFieldsFromMap(regularValues map[string]any) {
	delete(regularValues, "FathersName")
	delete(regularValues, "Category")
	delete(regularValues, "Address")
}

func createStudentWithDetails(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			regularValues, associationValues := splitAssociationValues(values)
			removeExtensionFieldsFromMap(regularValues)

			record := new(p_students.Student)
			err = db.Transaction(func(tx *gorm.DB) error {
				if err := baseviews.PopulateFromMap(record, regularValues); err != nil {
					return err
				}
				if err := tx.Create(record).Error; err != nil {
					return err
				}
				if err := applyAssociationReplacements(tx, record, associationValues); err != nil {
					return err
				}
				return upsertStudentDetails(tx, record.ID, values)
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

func updateStudentWithDetails(successURL getters.Getter[string]) func(*baseviews.View) http.Handler {
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
			regularValues, associationValues := splitAssociationValues(values)
			removeExtensionFieldsFromMap(regularValues)

			err = db.Transaction(func(tx *gorm.DB) error {
				query := tx.Model(new(p_students.Student)).Where("id = ?", id)
				for _, queryPatcher := range view.QueryPatchers {
					query = queryPatcher.Value(view, r, query)
				}

				record := new(p_students.Student)
				if err := query.First(record).Error; err != nil {
					return err
				}

				if len(regularValues) > 0 {
					updateQuery := tx.Model(new(p_students.Student)).Where("id = ?", id)
					for _, queryPatcher := range view.QueryPatchers {
						updateQuery = queryPatcher.Value(view, r, updateQuery)
					}
					if err := updateQuery.Updates(regularValues).Error; err != nil {
						return err
					}
				}

				if err := applyAssociationReplacements(tx, record, associationValues); err != nil {
					return err
				}

				return upsertStudentDetails(tx, uint(id), values)
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

func parseUintPathID(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	var id uint64
	// PathValue can return an empty string; treat it as invalid.
	if raw == "" {
		return 0, errors.New("empty id")
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	id = parsed
	return uint(id), nil
}

func QueryPatcherStudentDetailsStringContains(param, extensionField string) baseviews.QueryPatcher {
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

		detailFieldDBName, ok := fieldDBName[NirmancampusStudentDetails](query, extensionField)
		if !ok {
			return query
		}
		studentIDDBName, ok := fieldDBName[NirmancampusStudentDetails](query, "StudentID")
		if !ok {
			return query
		}

		subquery := query.Session(&gorm.Session{NewDB: true}).
			Model(new(NirmancampusStudentDetails)).
			Select(studentIDDBName).
			Where(detailFieldDBName+" ILIKE ?", "%"+value+"%")

		return query.Where("id IN (?)", subquery)
	}
}

func patchStudentViews() {
	const successDetail = "students.DetailRoute"

	successURL := lago.GetterRoutePath(successDetail, map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	})

	lago.RegistryView.Patch("students.CreateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = createStudentWithDetails(successURL)
		return view
	})

	lago.RegistryView.Patch("students.UpdateView", func(view *baseviews.View) *baseviews.View {
		view.Handlers[http.MethodPost] = updateStudentWithDetails(successURL)
		return view
	})

	lago.RegistryView.Patch("students.ListView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher("students.filter_fathers_name", QueryPatcherStudentDetailsStringContains("FathersName", "FathersName")).
			WithQueryPatcher("students.filter_category", QueryPatcherStudentDetailsStringContains("Category", "Category"))
	})

	lago.RegistryView.Patch("students.SelectView", func(view *baseviews.View) *baseviews.View {
		return view.WithQueryPatcher("students.filter_fathers_name", QueryPatcherStudentDetailsStringContains("FathersName", "FathersName")).
			WithQueryPatcher("students.filter_category", QueryPatcherStudentDetailsStringContains("Category", "Category"))
	})
}

func init() {
	patchStudentViews()
}

