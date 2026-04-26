package p_filesystem

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func currentVNodeTitle() getters.Getter[string] {
	return getters.Map(getters.Key[bool]("vnode.IsDirectory"), func(ctx context.Context, isDir bool) (string, error) {
		name, err := getters.Key[string]("vnode.Name")(ctx)
		if err != nil {
			return "", err
		}
		if isDir {
			return "Directory: " + name, nil
		}
		return "File: " + name, nil
	})
}

func currentVNodeBackRoute() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, _ := getters.Key[VNode]("vnode")(ctx)
		if n.ID == 0 {
			return lago.RoutePath("filesystem.ListRoute", nil)(ctx)
		}
		if n.ParentID == nil {
			return lago.RoutePath("filesystem.ListRoute", nil)(ctx)
		}
		return lago.RoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(*n.ParentID)),
		})(ctx)
	}
}

func listOrBrowseRoute(listRoute, browseRoute string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, _ := getters.Key[VNode]("vnode")(ctx)
		if n.ID == 0 {
			return lago.RoutePath(listRoute, nil)(ctx)
		}
		return lago.RoutePath(browseRoute, map[string]getters.Getter[any]{
			"parent_id": getters.Any(getters.Static(n.ID)),
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
		vn := VNode{FilePath: path}
		return vn.FileSizeDisplay(), nil
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
		n, err := getters.Key[VNode]("vnode")(ctx)
		if err != nil {
			return VNode{}, err
		}
		if n.ID == 0 || !n.IsDirectory {
			return VNode{}, fmt.Errorf("no current directory")
		}
		return n, nil
	}
}

func parentOfCurrentVNodeGetter() getters.Getter[VNode] {
	return func(ctx context.Context) (VNode, error) {
		p, err := getters.Key[VNode]("vnodeParent")(ctx)
		if err != nil {
			return p, err
		}
		if p.ID == 0 {
			return VNode{}, fmt.Errorf("no parent directory")
		}
		return p, nil
	}
}
