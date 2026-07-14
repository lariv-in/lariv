package views

import (
	"reflect"
	"testing"
)

type listFilterEnumStatus string

type listFilterTestModel struct {
	Name   string
	Status listFilterEnumStatus
	Count  int
}

func TestListFieldFilterMode(t *testing.T) {
	modelType := reflect.TypeOf(listFilterTestModel{})

	tests := []struct {
		field    string
		wantILIKE bool
	}{
		{field: "Name", wantILIKE: true},
		{field: "Status", wantILIKE: false},
		{field: "Count", wantILIKE: false},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			f, ok := modelType.FieldByName(tc.field)
			if !ok {
				t.Fatalf("field %q not found", tc.field)
			}
			gotILIKE := f.Type.Kind() == reflect.String && f.Type.Name() == "string"
			if gotILIKE != tc.wantILIKE {
				t.Fatalf("field %q: got ILIKE=%v, want %v (kind=%s name=%s)",
					tc.field, gotILIKE, tc.wantILIKE, f.Type.Kind(), f.Type.Name())
			}
		})
	}
}
