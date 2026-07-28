package getters

import (
	"context"
	"fmt"
)

// ContextKeyTableDisplay is set on the row context by table list/grid views so row-attribute
// getters can apply list vs grid styling ($tableDisplay is TableDisplayList or TableDisplayGrid).
const ContextKeyTableDisplay = "$tableDisplay"

const (
	TableDisplayList = "list"
	TableDisplayGrid = "grid"
)

func tableDisplayIsGrid(ctx context.Context) bool {
	v, _ := ctx.Value(ContextKeyTableDisplay).(string)
	return v == TableDisplayGrid
}

func rowAttrNavigateClick(click, classExpr Getter[string]) Getter[map[string]string] {
	return func(ctx context.Context) (map[string]string, error) {
		if click == nil {
			return nil, fmt.Errorf("getters: rowAttrNavigateClick: click getter is nil")
		}
		expr, err := click(ctx)
		if err != nil {
			return nil, err
		}
		var classStr string
		if classExpr != nil {
			classStr, err = classExpr(ctx)
			if err != nil {
				return nil, err
			}
		}
		grid := tableDisplayIsGrid(ctx)

		if expr == "" {
			if classStr != "" {
				return rowAttrClassOnly(classStr, grid), nil
			}
			return nil, nil
		}

		if grid {
			if classStr != "" {
				return map[string]string{
					"class":  "border border-base-300 rounded-box flex flex-col bg-base-100 p-2 cursor-pointer transition-colors",
					":class": classStr,
					"@click": expr,
				}, nil
			}
			return map[string]string{
				"class":  "border border-base-300 rounded-box flex flex-col bg-base-100 p-2 cursor-pointer hover:bg-base-200 transition-colors",
				"@click": expr,
			}, nil
		}

		if classStr != "" {
			return map[string]string{
				"class":  "cursor-pointer transition-colors",
				":class": classStr,
				"@click": expr,
			}, nil
		}
		return map[string]string{
			"class":  "cursor-pointer hover:bg-base-200 transition-colors",
			"@click": expr,
		}, nil
	}
}

func rowAttrClassOnly(classStr string, grid bool) map[string]string {
	if grid {
		return map[string]string{
			"class":  "border border-base-300 rounded-box flex flex-col bg-base-100 p-2 transition-colors",
			":class": classStr,
		}
	}
	return map[string]string{
		"class":  "transition-colors",
		":class": classStr,
	}
}
