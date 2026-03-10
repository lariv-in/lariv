package components

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Getter func(context.Context) any

func GetterStatic(value any) Getter {
	return func(ctx context.Context) any {
		return value
	}
}

func GetterKey(key string) Getter {
	return func(ctx context.Context) any {
		parts := strings.Split(key, ".")
		value := ctx.Value(parts[0])
		for i := 1; i < len(parts); i++ {
			if value == nil {
				return nil
			}
			m, ok := value.(map[string]any)
			if !ok {
				return nil
			}

			if v, exists := m[parts[i]]; exists {
				value = v
			} else if v, exists := m[strings.ToLower(parts[i])]; exists {
				value = v
			} else {
				found := false
				targetKey := strings.ToLower(strings.ReplaceAll(parts[i], "_", ""))
				for k, val := range m {
					if strings.ToLower(strings.ReplaceAll(k, "_", "")) == targetKey {
						value = val
						found = true
						break
					}
				}
				if !found {
					return nil
				}
			}
		}
		return value
	}
}

func GetterNil() Getter {
	return func(ctx context.Context) any {
		return nil
	}
}

func GetterFormat(format string, g ...Getter) Getter {
	return func(ctx context.Context) any {
		values := []any{}
		for _, getter := range g {
			values = append(values, IfOrGetter(getter, ctx, ""))
		}
		return fmt.Sprintf(format, values...)
	}
}

// Invokes the getter, if it is not nil and returns a non-nil value, returns that value. Otherwise returns the defaultValue.
func IfOrGetter(g Getter, ctx context.Context, defaultValue any) any {
	if g == nil {
		return defaultValue
	}
	value := g(ctx)
	if value == nil {
		return defaultValue
	}
	return value
}

// Invokes the getter, if it is not nil and returns a non-nil value, calls the builder. Otherwise returns the zero value of T.
func GetterIf[T any](g Getter, ctx context.Context, builder func(context.Context, any) T) T {
	var zero T
	if g == nil {
		return zero
	}
	value := g(ctx)
	if value == nil {
		return zero
	}
	return builder(ctx, value)
}

// GetterAssociation fetches a single record based on a foreign key dynamically at render time.
func GetterAssociation(table string, foreignKeyGetter Getter) Getter {
	return func(ctx context.Context) any {
		fkValue := foreignKeyGetter(ctx)
		if fkValue == nil || fkValue == "" {
			return nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return nil
		}

		result := map[string]any{}
		if err := db.Table(table).Where("id = ?", fkValue).Take(&result).Error; err != nil {
			return nil
		}
		return result
	}
}

// GetterNavigate returns an Alpine @click expression that performs HTMX navigation.
// urlFormat and getters work like GetterFormat to produce the URL per-row.
func GetterNavigate(urlFormat string, getters ...Getter) Getter {
	urlGetter := GetterFormat(urlFormat, getters...)
	return func(ctx context.Context) any {
		url := IfOrGetter(urlGetter, ctx, "")
		// Need to fix this so it uses htmx
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url)
	}
}

// GetterSelect returns an Alpine @click expression that dispatches an 'fk-select' event for single selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func GetterSelect(name string, valueGetter Getter, displayGetter Getter) Getter {
	return func(ctx context.Context) any {
		value := IfOrGetter(valueGetter, ctx, "")
		display := IfOrGetter(displayGetter, ctx, "")
		return fmt.Sprintf("$dispatch('fk-select',{name:'%s',value:'%v',display:'%v'})", name, value, display)
	}
}

// GetterMultiSelect returns an Alpine @click expression that dispatches an 'fk-multi-select' event for multi selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func GetterMultiSelect(name string, valueGetter Getter, displayGetter Getter) Getter {
	return func(ctx context.Context) any {
		value := IfOrGetter(valueGetter, ctx, "")
		display := IfOrGetter(displayGetter, ctx, "")
		return fmt.Sprintf("$dispatch('fk-multi-select',{name:'%s',value:'%v',display:'%v'})", name, value, display)
	}
}
