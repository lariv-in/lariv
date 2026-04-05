package getters

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
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

func rowAttrNavigateClick(click, classExpr Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
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
				return gomponents.Group{
					ghtml.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2 cursor-pointer transition-colors"),
					gomponents.Attr(":class", classStr),
					gomponents.Attr("@click", expr),
				}, nil
			}
			return gomponents.Group{
				ghtml.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2 cursor-pointer hover:bg-base-200 transition-colors"),
				gomponents.Attr("@click", expr),
			}, nil
		}

		if classStr != "" {
			return gomponents.Group{
				ghtml.Class("cursor-pointer transition-colors"),
				gomponents.Attr(":class", classStr),
				gomponents.Attr("@click", expr),
			}, nil
		}
		return gomponents.Group{
			ghtml.Class("cursor-pointer hover:bg-base-200 transition-colors"),
			gomponents.Attr("@click", expr),
		}, nil
	}
}

func rowAttrClassOnly(classStr string, grid bool) gomponents.Node {
	if grid {
		return gomponents.Group{
			ghtml.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2 transition-colors"),
			gomponents.Attr(":class", classStr),
		}
	}
	return gomponents.Group{
		ghtml.Class("transition-colors"),
		gomponents.Attr(":class", classStr),
	}
}
