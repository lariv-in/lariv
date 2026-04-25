package p_google_genai

import (
	"context"
	"reflect"
	"strings"
	"time"

	"google.golang.org/genai"
)

func NewClient(ctx context.Context) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: GoogleGenAIConfig.APIKey,
	})
}

func NewSchema[T any]() *genai.Schema {
	var schemaFromType func(reflect.Type) *genai.Schema
	schemaFromType = func(reflectType reflect.Type) *genai.Schema {
		if reflectType == nil {
			return &genai.Schema{Type: genai.TypeNULL}
		}

		nullable := false
		for reflectType.Kind() == reflect.Pointer {
			nullable = true
			reflectType = reflectType.Elem()
		}

		if reflectType == reflect.TypeFor[time.Time]() {
			schema := &genai.Schema{
				Type:   genai.TypeString,
				Format: "date-time",
			}
			if nullable {
				schema.Nullable = new(true)
			}
			return schema
		}

		var schema *genai.Schema
		switch reflectType.Kind() {
		case reflect.Bool:
			schema = &genai.Schema{Type: genai.TypeBoolean}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			schema = &genai.Schema{Type: genai.TypeInteger}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			schema = &genai.Schema{
				Type:    genai.TypeInteger,
				Minimum: new(float64(0)),
			}
		case reflect.Float32:
			schema = &genai.Schema{
				Type:   genai.TypeNumber,
				Format: "float",
			}
		case reflect.Float64:
			schema = &genai.Schema{
				Type:   genai.TypeNumber,
				Format: "double",
			}
		case reflect.String:
			schema = &genai.Schema{Type: genai.TypeString}
		case reflect.Array, reflect.Slice:
			if reflectType.Kind() == reflect.Slice && reflectType.Elem().Kind() == reflect.Uint8 {
				schema = &genai.Schema{
					Type:   genai.TypeString,
					Format: "byte",
				}
				break
			}

			schema = &genai.Schema{
				Type:  genai.TypeArray,
				Items: schemaFromType(reflectType.Elem()),
			}
		case reflect.Map:
			schema = &genai.Schema{Type: genai.TypeObject}
		case reflect.Struct:
			schema = newStructSchema(reflectType, schemaFromType)
		default:
			schema = &genai.Schema{Type: genai.TypeUnspecified}
		}

		if nullable {
			schema.Nullable = new(true)
		}

		return schema
	}

	return schemaFromType(reflect.TypeFor[T]())
}

func newStructSchema(reflectType reflect.Type, schemaFromType func(reflect.Type) *genai.Schema) *genai.Schema {
	schema := &genai.Schema{
		Type:             genai.TypeObject,
		Properties:       map[string]*genai.Schema{},
		PropertyOrdering: []string{},
	}
	requiredSeen := map[string]bool{}

	for _, field := range reflect.VisibleFields(reflectType) {
		name, options := parseJSONTag(field.Tag.Get("json"))
		if name == "-" {
			continue
		}

		if isPromotedStructField(field, name) || !field.IsExported() {
			continue
		}

		if name == "" {
			name = field.Name
		}

		fieldSchema := schemaFromType(field.Type)
		if options["string"] {
			fieldSchema = &genai.Schema{Type: genai.TypeString}
		}

		if _, ok := schema.Properties[name]; !ok {
			schema.PropertyOrdering = append(schema.PropertyOrdering, name)
		}
		schema.Properties[name] = fieldSchema

		if !options["omitempty"] && !requiredSeen[name] {
			requiredSeen[name] = true
			schema.Required = append(schema.Required, name)
		}
	}

	if len(schema.Properties) == 0 {
		schema.Properties = nil
		schema.PropertyOrdering = nil
	}
	if len(schema.Required) == 0 {
		schema.Required = nil
	}

	return schema
}

func isPromotedStructField(field reflect.StructField, name string) bool {
	if !field.Anonymous || name != "" {
		return false
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}

	return fieldType.Kind() == reflect.Struct
}

func parseJSONTag(tag string) (string, map[string]bool) {
	parts := strings.Split(tag, ",")
	options := map[string]bool{}
	for _, part := range parts[1:] {
		if part == "" {
			continue
		}
		options[part] = true
	}

	return parts[0], options
}
