package getters

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

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
