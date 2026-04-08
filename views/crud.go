package views

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/lariv-in/lago/components"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func PopulateFromMap[T any](v *T, values map[string]any) error {
	decodeConfig := mapstructure.DecoderConfig{Result: v, Deep: true, Squash: true, WeaklyTypedInput: true}
	decoder, err := mapstructure.NewDecoder(&decodeConfig)
	if err != nil {
		return err
	}
	return decoder.Decode(values)
}

func SplitAssociationValues(values map[string]any) (map[string]any, map[string]components.AssociationIDs) {
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

func applyAssociationReplacements[T any](db *gorm.DB, record *T, associations map[string]components.AssociationIDs) error {
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

func modelPrimaryKeyValue(record any) (any, error) {
	value := reflect.ValueOf(record)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, fmt.Errorf("record is nil")
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil, fmt.Errorf("record is invalid")
	}
	idField := value.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("record %T does not have an ID field", record)
	}
	switch idField.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return idField.Uint(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return idField.Int(), nil
	default:
		return nil, fmt.Errorf("record %T has unsupported ID field kind %s", record, idField.Kind())
	}
}

func uploadedJSONFile(values map[string]any, fileField string) (*multipart.FileHeader, error) {
	fileValue, ok := values[fileField]
	if !ok || fileValue == nil {
		return nil, fmt.Errorf("missing %q upload", fileField)
	}

	fileHeader, ok := fileValue.(*multipart.FileHeader)
	if !ok {
		return nil, fmt.Errorf("field %q must be a single file upload", fileField)
	}
	if fileHeader.Filename == "" {
		return nil, fmt.Errorf("field %q did not include a file", fileField)
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".json") {
		return nil, fmt.Errorf("field %q must be a .json file", fileField)
	}
	return fileHeader, nil
}

func decodeJSONArrayFile[T any](fileHeader *multipart.FileHeader) ([]T, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []T
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&records); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, fmt.Errorf("json upload must contain exactly one array")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != nil && err != io.EOF {
		return nil, err
	}
	if len(records) == 0 {
		return []T{}, nil
	}
	return records, nil
}

// sortQueryValueToOrder maps one URL sort value (e.g. "Name ASC", "User.Name DESC")
// to a GORM Order clause: each path segment is converted with the DB naming strategy
// (snake_case by default), and optional trailing ASC/DESC is preserved.
func sortQueryValueToOrder(namer schema.Namer, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ""
	}
	dir := ""
	colTokens := parts
	if n := len(parts); n >= 2 {
		last := strings.ToUpper(parts[n-1])
		if last == "ASC" || last == "DESC" {
			dir = " " + last
			colTokens = parts[:n-1]
		}
	}
	if len(colTokens) == 0 {
		return ""
	}
	ident := strings.Join(colTokens, " ")
	colExpr := sortIdentPathToSnakeColumns(namer, ident)
	if colExpr == "" {
		return ""
	}
	return colExpr + dir
}

func sortIdentPathToSnakeColumns(namer schema.Namer, ident string) string {
	segs := strings.Split(ident, ".")
	for i, s := range segs {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		segs[i] = namer.ColumnName("", s)
	}
	return strings.Join(segs, ".")
}
