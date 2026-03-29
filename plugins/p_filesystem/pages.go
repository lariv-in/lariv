package p_filesystem

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	registerMenus()
	registerFilters()
	registerForms()
	registerTables()
	registerDetail()
	registerSelection()
	registerDelete()
}

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
			return lago.GetterRoutePath("filesystem.ListRoute", nil)(ctx)
		}
		if node.ParentID == nil {
			return lago.GetterRoutePath("filesystem.ListRoute", nil)(ctx)
		}
		return lago.GetterRoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
			"parent_id": getters.GetterAny(getters.GetterStatic(*node.ParentID)),
		})(ctx)
	}
}

func currentVNodeDetailRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeEditRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.UpdateRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeDeleteRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.DeleteRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeMoveRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.MoveRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeBrowseRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
		"parent_id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeCreateChildRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.CreateChildRoute", map[string]getters.Getter[any]{
		"parent_id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeUploadChildRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.MultiUploadChildRoute", map[string]getters.Getter[any]{
		"parent_id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func currentVNodeDownloadRoute() getters.Getter[string] {
	return lago.GetterRoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
	})
}

func listOrBrowseRoute(listRoute, browseRoute string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		node, err := mustCurrentVNode(ctx)
		if err != nil {
			return lago.GetterRoutePath(listRoute, nil)(ctx)
		}
		return lago.GetterRoutePath(browseRoute, map[string]getters.Getter[any]{
			"parent_id": getters.GetterAny(getters.GetterStatic(node.ID)),
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
	return withSelectionTarget(lago.GetterRoutePath(childRoute, map[string]getters.Getter[any]{
		"parent_id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
	}))
}

func selectionRowClickGetter(defaultName string, modalID string, childRoute string, multi bool, selectDirectories bool) getters.Getter[string] {
	targetGetter := selectionTargetInput(defaultName)
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.GetterKey[bool]("$row.IsDirectory")(ctx)
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
			return getters.GetterMultiSelect(targetName,
				getters.GetterKey[uint]("$row.ID"),
				getters.GetterKey[string]("$row.Name"),
			)(ctx)
		}
		return getters.GetterSelect(targetName,
			getters.GetterKey[uint]("$row.ID"),
			getters.GetterKey[string]("$row.Name"),
		)(ctx)
	}
}

func vnodeTypeForKey(key string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.GetterKey[bool](key + ".IsDirectory")(ctx)
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
		isDirectory, err := getters.GetterKey[bool](key + ".IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		if isDirectory {
			return "-", nil
		}
		path, err := getters.GetterKey[string](key + ".FilePath")(ctx)
		if err != nil {
			return "", err
		}
		if path == "" {
			return "-", nil
		}
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "Missing", nil
			}
			return "Error", nil
		}
		return humanReadableSize(info.Size()), nil
	}
}

func vnodeChildrenCountForKey(key string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		isDirectory, err := getters.GetterKey[bool](key + ".IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		if !isDirectory {
			return "-", nil
		}
		id, err := getters.GetterKey[uint](key + ".ID")(ctx)
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
		isDirectory, err := getters.GetterKey[bool]("$row.IsDirectory")(ctx)
		if err != nil {
			return "", err
		}
		id, err := getters.GetterKey[uint]("$row.ID")(ctx)
		if err != nil {
			return "", err
		}
		if isDirectory {
			return lago.GetterRoutePath("filesystem.BrowseRoute", map[string]getters.Getter[any]{
				"parent_id": getters.GetterAny(getters.GetterStatic(id)),
			})(ctx)
		}
		return lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterStatic(id)),
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

func filesystemSidebar() []components.PageInterface {
	return []components.PageInterface{
		&components.ShowIf{
			Getter: currentVNodeExists(),
			Children: []components.PageInterface{
				lago.DynamicPage{Name: "filesystem.VNodeMenu"},
			},
		},
		&components.ShowIf{
			Getter: getters.GetterAny(func(ctx context.Context) (bool, error) {
				node, err := mustCurrentVNode(ctx)
				return err != nil || node.ID == 0, nil
			}),
			Children: []components.PageInterface{
				lago.DynamicPage{Name: "filesystem.MainMenu"},
			},
		},
	}
}

func vnodeFormFields(includeFileInput bool, submitLabel string, multi bool) components.ContainerColumn {
	initialIsDirectory := false
	isDirectoryGetter := getters.GetterKey[bool]("$in.IsDirectory")
	if v, err := isDirectoryGetter(context.Background()); err == nil {
		initialIsDirectory = v
	}

	children := []components.PageInterface{
		&components.ContainerError{
			Error: getters.GetterKey[error]("$error.ParentID"),
			Children: []components.PageInterface{
				&components.InputForeignKey[VNode]{
					Label:       "Location",
					Name:        "ParentID",
					Getter:      currentLocationGetter(),
					Url:         listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute"),
					Display:     getters.GetterKey[string]("$in.Name"),
					Placeholder: "Root (no parent)",
				},
			},
		},
		&components.ContainerError{
			Error: getters.GetterKey[error]("$error.IsDirectory"),
			Children: []components.PageInterface{
				&components.InputCheckbox{Label: "Create Directory", Name: "IsDirectory", Getter: getters.GetterKey[bool]("$in.IsDirectory"), XModel: "isDirectory"},
			},
		},
		&components.ClientIf{
			Condition: "isDirectory",
			Children: []components.PageInterface{&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$in.Name")},
				},
			}},
		},
	}

	if includeFileInput {
		label := "File"
		name := "File"
		if multi {
			label = "Files"
			name = "Files"
		}
		children = append(children, &components.ClientIf{
			Condition: "!isDirectory",
			Children: []components.PageInterface{&components.ContainerError{
				Error: getters.GetterKey[error]("$error." + name),
				Children: []components.PageInterface{
					&components.InputFile{Label: label, Name: name, Multiple: multi},
				},
			}},
		})
	}

	if submitLabel != "" {
		children = append(children, &components.ButtonSubmit{Label: submitLabel})
	}
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ClientData{
				Data:     fmt.Sprintf(`{ isDirectory: %t }`, initialIsDirectory),
				Children: children,
			},
		},
	}
}

func registerMenus() {
	lago.RegistryPage.Register("filesystem.MainMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Filesystem"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.GetterStatic("All Files"), Url: lago.GetterRoutePath("filesystem.ListRoute", nil), Icon: "folder-open"},
			&components.SidebarMenuItem{Title: getters.GetterStatic("Create Item"), Url: lago.GetterRoutePath("filesystem.CreateRoute", nil), Icon: "plus"},
			&components.SidebarMenuItem{Title: getters.GetterStatic("Bulk Upload"), Url: lago.GetterRoutePath("filesystem.MultiUploadRoute", nil), Icon: "arrow-up-tray"},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMenu", &components.SidebarMenu{
		Title: currentVNodeTitle(),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back"),
			Url:   currentVNodeBackRoute(),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.GetterStatic("View Details"), Url: currentVNodeDetailRoute(), Icon: "eye"},
			&components.SidebarMenuItem{Title: getters.GetterStatic("Edit"), Url: currentVNodeEditRoute(), Icon: "pencil-square"},
			&components.SidebarMenuItem{Title: getters.GetterStatic("Move"), Url: currentVNodeMoveRoute(), Icon: "arrow-right-circle"},
			&components.SidebarMenuItem{Title: getters.GetterStatic("Delete"), Url: currentVNodeDeleteRoute(), Icon: "trash"},
			&components.ShowIf{
				Getter: currentVNodeIsDirectory(),
				Children: []components.PageInterface{
					&components.SidebarMenuItem{Title: getters.GetterStatic("Browse Contents"), Url: currentVNodeBrowseRoute(), Icon: "folder-open"},
					&components.SidebarMenuItem{Title: getters.GetterStatic("Add New Item"), Url: currentVNodeCreateChildRoute(), Icon: "plus"},
					&components.SidebarMenuItem{Title: getters.GetterStatic("Bulk Upload"), Url: currentVNodeUploadChildRoute(), Icon: "arrow-up-tray"},
				},
			},
		},
	})
}

func registerFilters() {
	lago.RegistryPage.Register("filesystem.VNodeFilter", &components.FormComponent[VNode]{
		Url:    listOrBrowseRoute("filesystem.ListRoute", "filesystem.BrowseRoute"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("filesystem.ParentSelectionFilter", &components.FormComponent[VNode]{
		Url:    withSelectionTarget(listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute")),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("filesystem.DestinationSelectionFilter", &components.FormComponent[VNode]{
		Url:    withSelectionTarget(listOrBrowseRoute("filesystem.MoveSelectRoute", "filesystem.MoveSelectChildRoute")),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func registerForms() {
	lago.RegistryPage.Register("filesystem.VNodeCreateForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormComponent[VNode]{
				Url:      listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute"),
				Method:   http.MethodPost,
				Enctype:  "multipart/form-data",
				Title:    "Create Item",
				Subtitle: "Create a new file or directory",
				ChildrenInput: []components.PageInterface{
					vnodeFormFields(true, "Create", false),
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeUpdateForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormComponent[VNode]{
				Getter:   getters.GetterKey[VNode]("vnode"),
				Url:      currentVNodeEditRoute(),
				Method:   http.MethodPost,
				Enctype:  "multipart/form-data",
				Title:    "Edit Item",
				Subtitle: "Update file or directory details",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$in.Name"), Required: true},
						},
					},
					&components.ShowIf{
						Getter: getters.GetterAny(func(ctx context.Context) (bool, error) {
							isDirectory, err := getters.GetterKey[bool]("$in.IsDirectory")(ctx)
							if err != nil {
								return false, err
							}
							return !isDirectory, nil
						}),
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.GetterKey[error]("$error.File"),
								Children: []components.PageInterface{
									&components.InputFile{Label: "Replace File", Name: "File"},
								},
							},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save"},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMoveForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormComponent[VNode]{
				Getter:   getters.GetterKey[VNode]("vnode"),
				Url:      currentVNodeMoveRoute(),
				Method:   http.MethodPost,
				Title:    "Move Item",
				Subtitle: "Select the destination directory",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.DestinationID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[VNode]{
								Label:       "Destination Directory",
								Name:        "DestinationID",
								Getter:      parentOfCurrentVNodeGetter(),
								Url:         lago.GetterRoutePath("filesystem.MoveSelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Root (move to top level)",
							},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Move"},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMultiUploadForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormComponent[VNode]{
				Url:      listOrBrowseRoute("filesystem.MultiUploadRoute", "filesystem.MultiUploadChildRoute"),
				Method:   http.MethodPost,
				Enctype:  "multipart/form-data",
				Title:    "Bulk Upload",
				Subtitle: "Upload multiple files at once",
				ChildrenInput: []components.PageInterface{
					vnodeFormFields(false, "", false),
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Files"),
						Children: []components.PageInterface{
							&components.InputFile{Label: "Files", Name: "Files", Multiple: true, Required: true},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Upload Files"},
				},
			},
		},
	})
}

func registerTables() {
	lago.RegistryPage.Register("filesystem.VNodeTable", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.DataTable[VNode]{
				UID:             "filesystem-table",
				Data:            getters.GetterKey[components.ObjectList[VNode]]("vnodes"),
				Title:           "Filesystem",
				Subtitle:        "Files and folders",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "filesystem.VNodeFilter"}},
					&components.TableButtonCreate{Link: listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute")},
				},
				OnClick: getters.GetterNavigateGetter(rowOpenRoute()),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Size", Name: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$row")}}},
					{Label: "Items", Name: "Items", Children: []components.PageInterface{&components.FieldText{Getter: vnodeChildrenCountForKey("$row")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.UpdatedAt")}}},
				},
			},
		},
	})
}

func registerDetail() {
	lago.RegistryPage.Register("filesystem.VNodeDetail", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.Detail[VNode]{
				Getter: getters.GetterKey[VNode]("vnode"),
				Children: []components.PageInterface{
					&components.ContainerColumn{Children: []components.PageInterface{
						&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
						&components.FieldSubtitle{Getter: currentVNodeTitle()},
						&components.LabelInline{Title: "Path", Classes: "mt-2", Children: []components.PageInterface{&components.FieldText{Getter: currentVNodePath()}}},
						&components.LabelInline{Title: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$in")}}},
						&components.LabelInline{Title: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$in")}}},
						&components.LabelInline{Title: "Items", Children: []components.PageInterface{&components.FieldText{Getter: vnodeChildrenCountForKey("$in")}}},
						&components.LabelInline{Title: "Created At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.CreatedAt")}}},
						&components.LabelInline{Title: "Modified At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.UpdatedAt")}}},
						&components.ShowIf{
							Getter: getters.GetterAny(func(ctx context.Context) (bool, error) {
								isDirectory, err := getters.GetterKey[bool]("$in.IsDirectory")(ctx)
								if err != nil {
									return false, err
								}
								return !isDirectory, nil
							}),
							Children: []components.PageInterface{
								&components.ButtonDownload{Label: "Download", Link: lago.GetterRoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
									"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
								}), Icon: "arrow-down-tray", Classes: "btn-primary mt-4"},
							},
						},
					}},
				},
			},
		},
	})
}

func selectionTable(name string, filterName string, childRoute string, multi bool, selectDirectories bool) *components.Modal {
	title := "Select Directory"
	subtitle := "Choose a folder"
	if !selectDirectories {
		title = "Select Files"
		subtitle = "Choose files, or open folders to browse deeper"
	}

	modalID := "filesystem-selection-modal-" + name
	onClick := selectionRowClickGetter(name, modalID, childRoute, multi, selectDirectories)

	return &components.Modal{
		UID:   modalID,
		Title: title,
		Children: []components.PageInterface{
			&components.DataTable[VNode]{
				UID:             "filesystem-selection-table-" + name,
				Data:            getters.GetterKey[components.ObjectList[VNode]]("vnodes"),
				Title:           title,
				Subtitle:        subtitle,
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: filterName}},
				},
				OnClick: onClick,
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Path", Name: "Path", Children: []components.PageInterface{&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.UpdatedAt")}}},
				},
			},
		},
	}
}

func registerSelection() {
	lago.RegistryPage.Register("filesystem.ParentSelectionTable", selectionTable("ParentID", "filesystem.ParentSelectionFilter", "filesystem.SelectChildRoute", false, true))
	lago.RegistryPage.Register("filesystem.MultiSelectionTable", selectionTable("ParentID", "filesystem.ParentSelectionFilter", "filesystem.MultiSelectChildRoute", true, false))
	lago.RegistryPage.Register("filesystem.DestinationSelectionTable", selectionTable("DestinationID", "filesystem.DestinationSelectionFilter", "filesystem.MoveSelectChildRoute", false, true))
}

func registerDelete() {
	lago.RegistryPage.Register("filesystem.VNodeDeleteForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this item? Deleting directories will remove all nested contents.",
				CancelUrl: lago.GetterRoutePath("filesystem.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("vnode.ID")),
				}),
			},
		},
	})
}
