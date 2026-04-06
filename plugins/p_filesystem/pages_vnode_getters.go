package p_filesystem

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func mustCurrentVNode(ctx context.Context) (VNode, error) {
	raw := ctx.Value("vnode")
	switch node := raw.(type) {
	case VNode:
		if node.ID == 0 {
			return VNode{}, fmt.Errorf("missing current vnode")
		}
		return node, nil
	case *VNode:
		if node == nil || node.ID == 0 {
			return VNode{}, fmt.Errorf("missing current vnode")
		}
		return *node, nil
	default:
		return VNode{}, fmt.Errorf("missing current vnode")
	}
}

func currentVNodeExists() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		_, err := mustCurrentVNode(ctx)
		return err == nil, nil
	}
}

func currentVNodeIsDirectory() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return false, nil
		}
		return node.IsDirectory, nil
	}
}

func currentVNodeIsFile() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return false, nil
		}
		return !node.IsDirectory, nil
	}
}

func currentVNodeParentExists() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return false, nil
		}
		return node.ParentID != nil, nil
	}
}

func currentVNodeTitle() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return "", err
		}
		if node.IsDirectory {
			return fmt.Sprintf("Directory: %s", node.Name), nil
		}
		return fmt.Sprintf("File: %s", node.Name), nil
	}
}

func currentVNodePath() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return "", err
		}
		db, _ := ctx.Value("$db").(*gorm.DB)
		if db == nil {
			return "", fmt.Errorf("missing database in context")
		}
		return node.GetPath(db), nil
	}
}

func currentVNodeBackRoute() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return lago.RoutePath("filesystem.ListRoute", nil)(ctx)
		}
		if node.ParentID == nil {
			return lago.RoutePath("filesystem.ListRoute", nil)(ctx)
		}
		return lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(*node.ParentID)),
		})(ctx)
	}
}

func currentVNodeDetailRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeEditRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.UpdateRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeDeleteRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.DeleteRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeMoveRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.MoveRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeBrowseRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
		"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeCreateChildRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.CreateChildRoute", map[string]getters.Getter[any]{
		"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeUploadChildRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.MultiUploadChildRoute", map[string]getters.Getter[any]{
		"parent_id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func currentVNodeDownloadRoute() getters.Getter[string] {
	return lago.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Key[uint]("vnode.ID")),
	})
}

func listOrBrowseRoute(listRoute, browseRoute string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return lago.RoutePath(listRoute, nil)(ctx)
		}
		return lago.RoutePath(browseRoute, map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(node.ID)),
		})(ctx)
	}
}

func withSelectionTarget(routeGetter getters.Getter[string]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		route, err := routeGetter(ctx)
		if err != nil || route == "" {
			return route, err
		}
		r, _ := ctx.Value("$request").(*http.Request)
		if r == nil {
			return route, nil
		}
		targetInput := r.URL.Query().Get("target_input")
		if targetInput == "" {
			return route, nil
		}
		parsedURL, err := url.Parse(route)
		if err != nil {
			return route, nil
		}
		query := parsedURL.Query()
		query.Set("target_input", targetInput)
		parsedURL.RawQuery = query.Encode()
		return parsedURL.String(), nil
	}
}

func selectionTargetInput(defaultName string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		r, _ := ctx.Value("$request").(*http.Request)
		if r == nil {
			return defaultName, nil
		}
		if targetInput := r.URL.Query().Get("target_input"); targetInput != "" {
			return targetInput, nil
		}
		return defaultName, nil
	}
}

func selectionBrowseRouteGetter(childRoute string) getters.Getter[string] {
	return withSelectionTarget(lago.RoutePath(childRoute, map[string]getters.Getter[any]{
		"parent_id": getters.Any(getters.Key[uint]("$row.ID")),
	}))
}

func selectionRowClickGetter(defaultName, modalID, childRoute string, multi, selectDirectories bool) getters.Getter[string] {
	targetGetter := selectionTargetInput(defaultName)
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.Key[bool]("$row.IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		targetName, err := targetGetter(ctx)
		if err != nil {
			return "", err
		}

		if isDirectory && !selectDirectories {
			browseURL, err := selectionBrowseRouteGetter(childRoute)(ctx)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("htmx.ajax('GET', '%v', {target: '#%s', swap: 'outerHTML'})", browseURL, modalID), nil
		}

		if multi {
			return getters.SelectMulti(getters.Static(targetName),
				getters.Key[uint]("$row.ID"),
				getters.Key[string]("$row.Name"),
			)(ctx)
		}
		return getters.Select(targetName,
			getters.Key[uint]("$row.ID"),
			getters.Key[string]("$row.Name"),
		)(ctx)
	}
}

func vnodeTypeForKey(key string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.Key[bool](key + ".IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		if isDirectory {
			return "Directory", nil
		}
		return "File", nil
	}
}

func vnodeSizeForKey(key string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.Key[bool](key + ".IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		if isDirectory {
			return "-", nil
		}
		path, err := getters.Key[string](key + ".FilePath")(ctx)
		if err != nil {
			return "", err
		}
		if path == "" {
			return "-", nil
		}
		size, err := Store.StoredSize(path)
		if err != nil {
			if IsStoredFileMissing(err) {
				return "Missing", nil
			}
			return "Error", nil
		}
		return humanReadableSize(size), nil
	}
}

func vnodeChildrenCountForKey(key string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.Key[bool](key + ".IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		if !isDirectory {
			return "-", nil
		}
		id, err := getters.Key[uint](key + ".ID")(ctx)
		if err != nil {
			return "", err
		}
		db, _ := ctx.Value("$db").(*gorm.DB)
		if db == nil {
			return "", fmt.Errorf("missing database in context")
		}
		node := VNode{Model: gorm.Model{ID: id}, IsDirectory: true}
		return node.GetChildrenCount(db), nil
	}
}

func rowOpenRoute() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.Key[bool]("$row.IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		id, err := getters.Key[uint]("$row.ID")(ctx)
		if err != nil {
			return "", err
		}
		if isDirectory {
			return lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
				"parent_id": getters.Any(getters.Static(id)),
			})(ctx)
		}
		return lago.RoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(id)),
		})(ctx)
	}
}

func currentLocationGetter() getters.Getter[VNode] {
	return func(ctx context.Context) (VNode, error) {
		node, err := mustCurrentVNode(ctx)
		if err == nil && node.IsDirectory {
			return node, nil
		}
		var zero VNode
		return zero, fmt.Errorf("no current directory")
	}
}

func parentOfCurrentVNodeGetter() getters.Getter[VNode] {
	return func(ctx context.Context) (VNode, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil || node.ParentID == nil {
			var zero VNode
			return zero, fmt.Errorf("no parent directory")
		}
		db, _ := ctx.Value("$db").(*gorm.DB)
		if db == nil {
			var zero VNode
			return zero, fmt.Errorf("missing database in context")
		}
		parent, err := GetVNodeByID(db, *node.ParentID)
		if err != nil {
			var zero VNode
			return zero, err
		}
		return *parent, nil
	}
}
