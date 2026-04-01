package getters

import (
	"context"
	"reflect"
)

func associationIDsFromValue(raw any) []uint {
	if raw == nil {
		return nil
	}

	switch typed := raw.(type) {
	case []uint:
		return typed
	case []int:
		ids := make([]uint, 0, len(typed))
		for _, id := range typed {
			if id > 0 {
				ids = append(ids, uint(id))
			}
		}
		return ids
	}

	value := reflect.ValueOf(raw)
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Struct {
		idsField := value.FieldByName("IDs")
		if idsField.IsValid() {
			value = idsField
		}
	}

	if value.Kind() != reflect.Slice {
		return nil
	}

	ids := make([]uint, 0, value.Len())
	for i := range value.Len() {
		item := value.Index(i)
		for item.IsValid() && (item.Kind() == reflect.Pointer || item.Kind() == reflect.Interface) {
			if item.IsNil() {
				item = reflect.Value{}
				break
			}
			item = item.Elem()
		}
		if !item.IsValid() {
			continue
		}
		switch item.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ids = append(ids, uint(item.Uint()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if item.Int() > 0 {
				ids = append(ids, uint(item.Int()))
			}
		}
	}
	return ids
}

// AssociationIDs reads an association-ID style value from a context map
// such as $get and normalizes it into []uint.
func AssociationIDs(contextKey, field string) Getter[[]uint] {
	return func(ctx context.Context) ([]uint, error) {
		value := ctx.Value(contextKey)
		if value == nil {
			return nil, nil
		}
		getMap, ok := value.(map[string]any)
		if !ok {
			return nil, nil
		}
		raw, ok := getMap[field]
		if !ok {
			return nil, nil
		}
		return associationIDsFromValue(raw), nil
	}
}
