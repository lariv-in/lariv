package p_google_genai

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/genai"
)

type schemaEmbedded struct {
	Embedded string `json:"embedded"`
}

type schemaSample struct {
	schemaEmbedded
	Name     string         `json:"name"`
	Optional *int           `json:"optional,omitempty"`
	Bytes    []byte         `json:"bytes,omitempty"`
	When     time.Time      `json:"when"`
	Count    int            `json:"count,string"`
	Meta     map[string]any `json:"meta,omitempty"`
	Text     string         `json:"text"`
	Ignored  string         `json:"-"`
}

type badMap map[float64]string

func TestNewSchemaBasicKinds(t *testing.T) {
	if schema := NewSchema[bool](); schema.Type != genai.TypeBoolean {
		t.Fatalf("bool schema type = %v, want %v", schema.Type, genai.TypeBoolean)
	}

	if schema := NewSchema[uint](); schema.Type != genai.TypeInteger || schema.Minimum == nil || *schema.Minimum != 0 {
		t.Fatalf("uint schema = %#v, want integer with minimum 0", schema)
	}

	if schema := NewSchema[float32](); schema.Type != genai.TypeNumber || schema.Format != "float" {
		t.Fatalf("float32 schema = %#v, want number/float", schema)
	}

	if schema := NewSchema[float64](); schema.Type != genai.TypeNumber || schema.Format != "double" {
		t.Fatalf("float64 schema = %#v, want number/double", schema)
	}

	if schema := NewSchema[[]byte](); schema.Type != genai.TypeString || schema.Format != "byte" {
		t.Fatalf("[]byte schema = %#v, want string/byte", schema)
	}

	if schema := NewSchema[[]string](); schema.Type != genai.TypeArray || schema.Items == nil || schema.Items.Type != genai.TypeString {
		t.Fatalf("[]string schema = %#v, want array of strings", schema)
	}

	if schema := NewSchema[map[string]int](); schema.Type != genai.TypeObject {
		t.Fatalf("map[string]int schema type = %v, want %v", schema.Type, genai.TypeObject)
	}

	if schema := NewSchema[any](); schema.Type != genai.TypeUnspecified {
		t.Fatalf("any schema type = %v, want %v", schema.Type, genai.TypeUnspecified)
	}

	if schema := NewSchema[func()](); schema.Type != genai.TypeUnspecified {
		t.Fatalf("func schema type = %v, want %v", schema.Type, genai.TypeUnspecified)
	}
}

func TestNewSchemaStructFields(t *testing.T) {
	schema := NewSchema[schemaSample]()
	if schema.Type != genai.TypeObject {
		t.Fatalf("struct schema type = %v, want %v", schema.Type, genai.TypeObject)
	}

	required := []string{"embedded", "name", "when", "count", "text"}
	if !reflect.DeepEqual(schema.Required, required) {
		t.Fatalf("required = %v, want %v", schema.Required, required)
	}

	if _, ok := schema.Properties["Ignored"]; ok {
		t.Fatal("ignored field should not be present")
	}

	if got := schema.Properties["embedded"]; got == nil || got.Type != genai.TypeString {
		t.Fatalf("embedded schema = %#v, want string", got)
	}

	if got := schema.Properties["optional"]; got == nil || got.Nullable == nil || !*got.Nullable {
		t.Fatalf("optional schema = %#v, want nullable", got)
	}

	if got := schema.Properties["bytes"]; got == nil || got.Type != genai.TypeString || got.Format != "byte" {
		t.Fatalf("bytes schema = %#v, want string/byte", got)
	}

	if got := schema.Properties["when"]; got == nil || got.Type != genai.TypeString || got.Format != "date-time" {
		t.Fatalf("when schema = %#v, want string/date-time", got)
	}

	if got := schema.Properties["count"]; got == nil || got.Type != genai.TypeString {
		t.Fatalf("count schema = %#v, want string because of ,string tag", got)
	}

	if got := schema.Properties["meta"]; got == nil || got.Type != genai.TypeObject {
		t.Fatalf("meta schema = %#v, want object", got)
	}

	if got := schema.Properties["text"]; got == nil || got.Type != genai.TypeString {
		t.Fatalf("text schema = %#v, want string", got)
	}
}

func TestNewSchemaMapKeyTypeIgnored(t *testing.T) {
	schema := NewSchema[badMap]()
	if schema.Type != genai.TypeObject {
		t.Fatalf("bad map schema type = %v, want %v", schema.Type, genai.TypeObject)
	}
}
