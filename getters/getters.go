package getters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Context key constants for shared use across packages.
const (
	ContextKeyError = "$error"
	ContextKeyGet   = "$get"
	ContextKeyIn    = "$in"
)

// Getter defines a common type for fetching data that could be dynamic
type Getter[T any] func(context.Context) (T, error)

// Static returns a Getter which will always return a static value
// Never errors
func Static[T any](value T) Getter[T] {
	return func(ctx context.Context) (T, error) {
		return value, nil
	}
}

// Key returns a Getter that gets the value from the context.
// '.' can be used to traverse map or struct fields. Keys must match exactly.
// Returns the zero value of T when key is not found, with an error
func Key[T any](key string) Getter[T] {
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
			} else {
				value = nil
			}
		}
		if value == nil {
			return zero, nil
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

func NumberCast[T Number, V Number](g Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return T(value), nil
	}
}

func QueryEscape[T comparable](g Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		value, err := IfOr(g, ctx, zero)
		if err != nil {
			return "", err
		}
		return url.QueryEscape(fmt.Sprintf("%v", value)), nil
	}
}

func Nil[T any]() Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		return zero, nil
	}
}

func Any[T any](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		return g(ctx)
	}
}

// BoolNot negates a boolean getter. The result is Getter[any] for use with ShowIf and similar.
func BoolNot[T ~bool](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		v, err := g(ctx)
		if err != nil {
			return nil, err
		}
		return !v, nil
	}
}

// IntString converts a Getter[int] to Getter[string] by formatting the int.
// Errors from the underlying getter (e.g. type mismatch) are propagated.
func IntString(g Getter[int]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(v), nil
	}
}

// UintString converts a Getter[uint] to Getter[string] by formatting the uint.
// Errors from the underlying getter are propagated.
func UintString(g Getter[uint]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(uint64(v), 10), nil
	}
}

// TimeFormat converts a Getter[time.Time] to Getter[string] by formatting
// the time using the provided layout. Errors from the underlying getter are
// propagated.
func TimeFormat(layout string, g Getter[time.Time]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := g(ctx)
		if err != nil {
			return "", err
		}
		if t.IsZero() {
			return "", nil
		}
		return t.Format(layout), nil
	}
}

func Format(format string, g ...Getter[any]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		values := []any{}
		for _, getter := range g {
			v, err := IfOr(getter, ctx, "")
			if err != nil {
				return "", err
			}
			values = append(values, v)
		}
		return fmt.Sprintf(format, values...), nil
	}
}

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

// ContextAssociationIDs reads an association-ID style value from a context map
// such as $get and normalizes it into []uint.
func ContextAssociationIDs(contextKey, field string) Getter[[]uint] {
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

// Invokes the getter, if it is not nil and returns a non-nil value, and does not error out, returns that value. Otherwise returns the defaultValue.
func IfOr[T comparable](g Getter[T], ctx context.Context, defaultValue T) (T, error) {
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

// IfOrElse returns a Getter that invokes g when g is non-nil and returns a non-zero value without error;
// otherwise it invokes elseGetter. If elseGetter is nil in those fallback cases, returns the zero value of T.
func IfOrElse[T comparable](g Getter[T], elseGetter Getter[T]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		if g != nil {
			value, err := g(ctx)
			if err == nil && value != zero {
				return value, nil
			}
		}
		if elseGetter != nil {
			return elseGetter(ctx)
		}
		return zero, nil
	}
}

// Invokes the getter, if it is not nil and returns a non-nil value and does not error out, calls the builder. Otherwise returns the zero value of T.
func If[T any, V comparable](g Getter[V], ctx context.Context, builder func(context.Context, V) (T, error)) (T, error) {
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

// Association fetches a single record based on a foreign key dynamically at render time.
func Association[T any, V any](foreignKeyGetter Getter[V]) Getter[T] {
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
		if err := db.Model(new(T)).Where("id = ?", fkValue).Take(&result).Error; err != nil {
			return zero, err
		}
		return result, nil
	}
}

// ForeignKey fetches a related model T by its primary key and returns a specific field.
// foreignKeyGetter resolves the FK value (e.g. GetterKey("$in.RoleID")).
// fieldPath is the dot-separated path into the related model's map (e.g. "Name").
func ForeignKey[T any, K comparable, V any](foreignKeyGetter Getter[K], fieldPath string) Getter[V] {
	var zeroK K
	var zeroV V
	return func(ctx context.Context) (V, error) {
		fkValue, err := IfOr(foreignKeyGetter, ctx, zeroK)
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

func schemaFieldDBName[T any](db *gorm.DB, fieldName string) (string, error) {
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

func idMapForSlice[T any](items []T) map[uint]T {
	byID := make(map[uint]T, len(items))
	for _, item := range items {
		valueMap := MapFromStruct(item)
		rawID, ok := valueMap["ID"]
		if !ok {
			continue
		}
		switch typed := rawID.(type) {
		case uint:
			byID[typed] = item
		case uint8:
			byID[uint(typed)] = item
		case uint16:
			byID[uint(typed)] = item
		case uint32:
			byID[uint(typed)] = item
		case uint64:
			byID[uint(typed)] = item
		case int:
			if typed > 0 {
				byID[uint(typed)] = item
			}
		case int8:
			if typed > 0 {
				byID[uint(typed)] = item
			}
		case int16:
			if typed > 0 {
				byID[uint(typed)] = item
			}
		case int32:
			if typed > 0 {
				byID[uint(typed)] = item
			}
		case int64:
			if typed > 0 {
				byID[uint(typed)] = item
			}
		}
	}
	return byID
}

func orderedAssociationSlice[T any](items []T, ids []uint, order string) []T {
	if order != "" {
		return items
	}
	byID := idMapForSlice(items)
	ordered := make([]T, 0, len(ids))
	for _, id := range ids {
		if item, ok := byID[id]; ok {
			ordered = append(ordered, item)
		}
	}
	return ordered
}

// AssociationList fetches multiple records by ID and returns them as a slice.
// When order is empty, the returned slice preserves the order of idsGetter.
func AssociationList[T any](idsGetter Getter[[]uint], order string, preloads ...string) Getter[[]T] {
	return func(ctx context.Context) ([]T, error) {
		ids, err := idsGetter(ctx)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return nil, errors.New("Couldn't load db connection from context")
		}

		query := db.Model(new(T))
		for _, preload := range preloads {
			query = query.Preload(preload)
		}
		if order != "" {
			query = query.Order(order)
		}

		var results []T
		if err := query.Where("id IN ?", ids).Find(&results).Error; err != nil {
			return nil, err
		}
		return orderedAssociationSlice(results, ids, order), nil
	}
}

// JoinAssociationList fetches related records through a join model.
// ownerField and targetField are struct field names on the join model (for example,
// "CourseID" and "TeacherID"). When order is empty, join-row order is preserved.
func JoinAssociationList[TJoin any, TTarget any](ownerIDGetter Getter[uint], ownerField, targetField, order string, preloads ...string) Getter[[]TTarget] {
	return func(ctx context.Context) ([]TTarget, error) {
		ownerID, err := ownerIDGetter(ctx)
		if err != nil {
			return nil, err
		}
		if ownerID == 0 {
			return nil, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return nil, errors.New("Couldn't load db connection from context")
		}

		ownerDBName, err := schemaFieldDBName[TJoin](db, ownerField)
		if err != nil {
			return nil, err
		}
		targetDBName, err := schemaFieldDBName[TJoin](db, targetField)
		if err != nil {
			return nil, err
		}

		joinQuery := db.Model(new(TJoin)).Where(ownerDBName+" = ?", ownerID)
		if order == "" {
			joinQuery = joinQuery.Order(ownerDBName + " ASC")
		}

		var targetIDs []uint
		if err := joinQuery.Pluck(targetDBName, &targetIDs).Error; err != nil {
			return nil, err
		}
		if len(targetIDs) == 0 {
			return nil, nil
		}

		return AssociationList[TTarget](Static(targetIDs), order, preloads...)(ctx)
	}
}

// Navigate returns an Alpine @click expression that performs HTMX navigation.
// urlFormat and getters work like GetterFormat to produce the URL per-row.
func Navigate(urlFormat string, getters ...Getter[any]) Getter[string] {
	urlGetter := Format(urlFormat, getters...)
	return func(ctx context.Context) (string, error) {
		url, err := IfOr(urlGetter, ctx, "")
		if err != nil {
			return "", err
		}
		// Need to fix this so it uses htmx
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}

// NavigateGetter is like GetterNavigate but takes a pre-built Getter for the URL.
func NavigateGetter[T comparable](urlGetter Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		url, err := IfOr(urlGetter, ctx, zero)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}

// Select returns an Alpine @click expression that dispatches an 'fk-select' event for single selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func Select[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOr(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("$dispatch('fk-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}

// SelectMulti returns an Alpine @click expression that dispatches an
// 'fk-multi-select' event for multi-selection inputs. nameGetter resolves the
// input field name (e.g. getters.GetterStatic("Field"), or
// getters.IfOrElseGetter(getters.GetterKey[string]("$get.target_input"), getters.GetterStatic("Field"))
// when the name may come from target_input on the request).
func SelectMulti[T, D comparable](nameGetter Getter[string], valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		name, err := IfOr(nameGetter, ctx, "")
		if err != nil {
			return "", err
		}
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOr(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("$dispatch('fk-multi-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}

func Deref[T any](g Getter[*T]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		if value == nil {
			return zero, nil
		}
		return *value, nil
	}
}

func Map[T, V any](g Getter[T], f func(context.Context, T) (V, error)) Getter[V] {
	var zero V
	return func(ctx context.Context) (V, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return f(ctx, value)
	}
}

// JSONObj parses JSON from a string getter and returns datatypes.JSON. The top-level value must be a JSON object.
// Empty or whitespace-only input yields "{}".
func JSONObj[T any](g Getter[string]) Getter[datatypes.JSON] {
	return func(ctx context.Context) (datatypes.JSON, error) {
		s, err := g(ctx)
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return datatypes.JSON("{}"), nil
		}
		var m T
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, err
		}
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return datatypes.JSON(b), nil
	}
}

// JSONArray parses JSON from a string getter and returns datatypes.JSON. The top-level value must be a JSON array.
// Empty or whitespace-only input yields "[]".
func JSONArray[T any](g Getter[string]) Getter[datatypes.JSON] {
	return func(ctx context.Context) (datatypes.JSON, error) {
		s, err := g(ctx)
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return datatypes.JSON("[]"), nil
		}
		var a []T
		if err := json.Unmarshal([]byte(s), &a); err != nil {
			return nil, err
		}
		b, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		return datatypes.JSON(b), nil
	}
}

func ParseUint(g Getter[string]) Getter[uint] {
	return func(ctx context.Context) (uint, error) {
		s, err := g(ctx)
		if err != nil {
			return 0, err
		}
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(u), nil
	}
}

func ParseInt(g Getter[string]) Getter[int] {
	return func(ctx context.Context) (int, error) {
		s, err := g(ctx)
		if err != nil {
			return 0, err
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i), nil
	}
}
