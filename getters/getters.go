package getters

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
	"gorm.io/gorm"
)

// Context key constants for shared use across packages.
const (
	ContextKeyError = "$error"
	ContextKeyIn    = "$in"
)

// Getter defines a common type for fetching data that could be dynamic
type Getter[T any] func(context.Context) (T, error)

// GetterStatic returns a Getter which will always return a static value
// Never errors
func GetterStatic[T any](value T) Getter[T] {
	return func(ctx context.Context) (T, error) {
		return value, nil
	}
}

// GetterKey returns a Getter that gets the value from the context.
// '.' can be used to traverse map or struct fields. Keys must match exactly.
// Returns the zero value of T when key is not found, with an error
func GetterKey[T any](key string) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		parts := strings.Split(key, ".")
		value := ctx.Value(parts[0])
		for _, part := range parts[1:] {
			if value == nil {
				return zero, fmt.Errorf("Couldn't find %s in context", key)
			}
			m, ok := value.(map[string]any)
			if !ok {
				v, ok := value.(reflect.Value)
				if !ok {
					v = reflect.ValueOf(value)
				}

				if v.Kind() == reflect.Pointer {
					v = v.Elem()
				}
				m = MapFromStruct(v)
			}

			if v, exists := m[part]; exists {
				value = v
			}
		}
		v, ok := value.(T)
		if !ok {
			return zero, fmt.Errorf("Value for key %s found, but the type of value in context was %v, expected %v", key, reflect.TypeOf(value), reflect.TypeOf(zero))
		}
		return v, nil
	}
}

type Number interface {
	constraints.Integer | constraints.Float
}

func GetterNumberCast[T Number, V Number](g Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return T(value), nil
	}
}

func GetterQueryEscape[T comparable](g Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		value, err := IfOrGetter(g, ctx, zero)
		if err != nil {
			return "", err
		}
		return url.QueryEscape(fmt.Sprintf("%v", value)), nil
	}
}

func GetterNil[T any]() Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		return zero, nil
	}
}

func GetterAny[T any](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		return g(ctx)
	}
}

// GetterIntString converts a Getter[int] to Getter[string] by formatting the int.
// Errors from the underlying getter (e.g. type mismatch) are propagated.
func GetterIntString(g Getter[int]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(v), nil
	}
}

// GetterUintString converts a Getter[uint] to Getter[string] by formatting the uint.
// Errors from the underlying getter are propagated.
func GetterUintString(g Getter[uint]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(uint64(v), 10), nil
	}
}

func GetterFormat(format string, g ...Getter[any]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		values := []any{}
		for _, getter := range g {
			v, err := IfOrGetter(getter, ctx, "")
			if err != nil {
				return "", err
			}
			values = append(values, v)
		}
		return fmt.Sprintf(format, values...), nil
	}
}

// Invokes the getter, if it is not nil and returns a non-nil value, and does not error out, returns that value. Otherwise returns the defaultValue.
func IfOrGetter[T comparable](g Getter[T], ctx context.Context, defaultValue T) (T, error) {
	var zero T
	if g == nil {
		return defaultValue, nil
	}
	value, err := g(ctx)
	if err != nil {
		return defaultValue, nil
	}
	if value == zero {
		return defaultValue, nil
	}
	return value, nil
}

// Invokes the getter, if it is not nil and returns a non-nil value and does not error out, calls the builder. Otherwise returns the zero value of T.
func GetterIf[T any, V comparable](g Getter[V], ctx context.Context, builder func(context.Context, V) (T, error)) (T, error) {
	var zero T
	var zeroV V
	if g == nil {
		return zero, errors.New("Getter is nil")
	}
	value, err := g(ctx)
	if err != nil {
		return zero, err
	}
	if value == zeroV {
		return zero, errors.New("Value is nil")
	}
	return builder(ctx, value)
}

// GetterAssociation fetches a single record based on a foreign key dynamically at render time.
func GetterAssociation[T any, V any](table string, foreignKeyGetter Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		fkValue, err := foreignKeyGetter(ctx)
		if err != nil {
			return zero, err
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return zero, errors.New("Couldn't load db connection from context")
		}

		var result T
		if err := db.Table(table).Where("id = ?", fkValue).Take(&result).Error; err != nil {
			return zero, err
		}
		return result, nil
	}
}

// GetterForeignKey fetches a related model T by its primary key and returns a specific field.
// foreignKeyGetter resolves the FK value (e.g. GetterKey("$in.RoleID")).
// fieldPath is the dot-separated path into the related model's map (e.g. "Name").
func GetterForeignKey[T any, K comparable, V any](foreignKeyGetter Getter[K], fieldPath string) Getter[V] {
	var zeroK K
	var zeroV V
	return func(ctx context.Context) (V, error) {
		fkValue, err := IfOrGetter(foreignKeyGetter, ctx, zeroK)
		if err != nil {
			return zeroV, err
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return zeroV, errors.New("Couldn't load db connection from context")
		}

		var instance T
		if err := db.First(&instance, fkValue).Error; err != nil {
			return zeroV, err
		}

		// Convert to map and walk the field path
		m := MapFromStruct(&instance)
		parts := strings.Split(fieldPath, ".")
		var value any = m
		for _, part := range parts {
			mp, ok := value.(map[string]any)
			if !ok {
				return zeroV, errors.New("Couldn't convert the related field struct to map")
			}
			value, ok = mp[part]
			if !ok {
				return zeroV, errors.New("Couldn't find the key in the struct")
			}
		}
		v, ok := value.(V)
		if !ok {
			return zeroV, fmt.Errorf("Value for key %s found, but the type of value in context was %v, expected %v", fieldPath, reflect.TypeOf(value), reflect.TypeOf(zeroV))
		}
		return v, nil
	}
}

// GetterNavigate returns an Alpine @click expression that performs HTMX navigation.
// urlFormat and getters work like GetterFormat to produce the URL per-row.
func GetterNavigate(urlFormat string, getters ...Getter[any]) Getter[string] {
	urlGetter := GetterFormat(urlFormat, getters...)
	return func(ctx context.Context) (string, error) {
		url, err := IfOrGetter(urlGetter, ctx, "")
		if err != nil {
			return "", err
		}
		// Need to fix this so it uses htmx
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}

// GetterNavigateGetter is like GetterNavigate but takes a pre-built Getter for the URL.
func GetterNavigateGetter[T comparable](urlGetter Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		url, err := IfOrGetter(urlGetter, ctx, zero)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}

// GetterSelect returns an Alpine @click expression that dispatches an 'fk-select' event for single selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func GetterSelect[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		value, err := IfOrGetter(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOrGetter(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("$dispatch('fk-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}

// GetterMultiSelect returns an Alpine @click expression that dispatches an 'fk-multi-select' event for multi selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func GetterMultiSelect[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		value, err := IfOrGetter(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOrGetter(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("$dispatch('fk-multi-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}
