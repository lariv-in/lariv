package p_filesystem

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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

func filesystemSidebar() []components.PageInterface {
	return []components.PageInterface{
		&components.ShowIf{
			Getter: currentVNodeExists(),
			Children: []components.PageInterface{
				lago.DynamicPage{Name: "filesystem.VNodeMenu"},
			},
		},
		&components.ShowIf{
			Getter: getters.Any(func(ctx context.Context) (bool, error) {
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
	isDirectoryGetter := getters.Key[bool]("$in.IsDirectory")
	if v, err := isDirectoryGetter(context.Background()); err == nil {
		initialIsDirectory = v
	}

	children := []components.PageInterface{
		&components.ContainerError{
			Error: getters.Key[error]("$error.ParentID"),
			Children: []components.PageInterface{
				&components.InputForeignKey[VNode]{
					Label:       "Location",
					Name:        "ParentID",
					Getter:      currentLocationGetter(),
					Url:         listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute"),
					Display:     getters.Key[string]("$in.Name"),
					Placeholder: "Root (no parent)",
				},
			},
		},
		&components.ContainerError{
			Error: getters.Key[error]("$error.IsDirectory"),
			Children: []components.PageInterface{
				&components.InputCheckbox{Label: "Create Directory", Name: "IsDirectory", Getter: getters.Key[bool]("$in.IsDirectory"), XModel: "isDirectory"},
			},
		},
		&components.ClientIf{
			Condition: "isDirectory",
			Children: []components.PageInterface{&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$in.Name")},
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
				Error: getters.Key[error]("$error." + name),
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
		Title: getters.Static("Filesystem"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("All Files"), Url: lago.RoutePath("filesystem.ListRoute", nil), Icon: "folder-open"},
			&components.SidebarMenuItem{Title: getters.Static("Create Item"), Url: lago.RoutePath("filesystem.CreateRoute", nil), Icon: "plus"},
			&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: lago.RoutePath("filesystem.MultiUploadRoute", nil), Icon: "arrow-up-tray"},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMenu", &components.SidebarMenu{
		Title: currentVNodeTitle(),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back"),
			Url:   currentVNodeBackRoute(),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("View Details"), Url: currentVNodeDetailRoute(), Icon: "eye"},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: currentVNodeEditRoute(), Icon: "pencil-square"},
			&components.SidebarMenuItem{Title: getters.Static("Move"), Url: currentVNodeMoveRoute(), Icon: "arrow-right-circle"},
			&components.ShowIf{
				Getter: currentVNodeIsDirectory(),
				Children: []components.PageInterface{
					&components.SidebarMenuItem{Title: getters.Static("Browse Contents"), Url: currentVNodeBrowseRoute(), Icon: "folder-open"},
					&components.SidebarMenuItem{Title: getters.Static("Add New Item"), Url: currentVNodeCreateChildRoute(), Icon: "plus"},
					&components.SidebarMenuItem{Title: getters.Static("Bulk Upload"), Url: currentVNodeUploadChildRoute(), Icon: "arrow-up-tray"},
				},
			},
		},
	})
}

func registerFilters() {
	lago.RegistryPage.Register("filesystem.VNodeFilter", &components.FormComponent[VNode]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(listOrBrowseRoute("filesystem.ListRoute", "filesystem.BrowseRoute"))),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("filesystem.ParentSelectionFilter", &components.FormComponent[VNode]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(withSelectionTarget(listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute")))),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("filesystem.DestinationSelectionFilter", &components.FormComponent[VNode]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(withSelectionTarget(listOrBrowseRoute("filesystem.MoveSelectRoute", "filesystem.MoveSelectChildRoute")))),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
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
				Attr:     getters.FormAttr(http.MethodPost, getters.FormSubmit(listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute"))),
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
				Getter:   getters.Key[VNode]("vnode"),
				Attr:     getters.FormAttr(http.MethodPost, getters.FormSubmit(currentVNodeEditRoute())),
				Title:    "Edit Item",
				Subtitle: "Update file or directory details",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$in.Name"), Required: true},
						},
					},
					&components.ShowIf{
						Getter: getters.Any(func(ctx context.Context) (bool, error) {
							isDirectory, err := getters.Key[bool]("$in.IsDirectory")(ctx)
							if err != nil {
								return false, err
							}
							return !isDirectory, nil
						}),
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.File"),
								Children: []components.PageInterface{
									&components.InputFile{Label: "Replace File", Name: "File"},
								},
							},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							&components.ButtonModal{
								Label:   "Delete",
								Icon:    "trash",
								Url:     currentVNodeDeleteRoute(),
								Classes: "btn-outline btn-error btn-sm",
							},
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMoveForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormComponent[VNode]{
				Getter: getters.Key[VNode]("vnode"),
				Attr:   getters.FormAttr(http.MethodPost, getters.FormSubmit(currentVNodeMoveRoute())),

				Title:    "Move Item",
				Subtitle: "Select the destination directory",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.DestinationID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[VNode]{
								Label:       "Destination Directory",
								Name:        "DestinationID",
								Getter:      parentOfCurrentVNodeGetter(),
								Url:         lago.RoutePath("filesystem.MoveSelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
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
				Attr:     getters.FormAttr(http.MethodPost, getters.FormSubmit(listOrBrowseRoute("filesystem.MultiUploadRoute", "filesystem.MultiUploadChildRoute"))),
				Title:    "Bulk Upload",
				Subtitle: "Upload multiple files at once",
				ChildrenInput: []components.PageInterface{
					vnodeFormFields(false, "", false),
					&components.ContainerError{
						Error: getters.Key[error]("$error.Files"),
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
				UID:      "filesystem-table",
				Data:     getters.Key[components.ObjectList[VNode]]("vnodes"),
				Title:    "Filesystem",
				Subtitle: "Files and folders",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "filesystem.VNodeFilter"}},
					&components.TableButtonCreate{Link: listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute")},
				},
				RowAttr: getters.RowAttrNavigate(rowOpenRoute()),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Size", Name: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$row")}}},
					{Label: "Items", Name: "Items", Children: []components.PageInterface{&components.FieldText{Getter: vnodeChildrenCountForKey("$row")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")}}},
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
				Getter: getters.Key[VNode]("vnode"),
				Children: []components.PageInterface{
					&components.ContainerColumn{Children: []components.PageInterface{
						&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
						&components.FieldSubtitle{Getter: currentVNodeTitle()},
						&components.LabelInline{Title: "Path", Classes: "mt-2", Children: []components.PageInterface{&components.FieldText{Getter: currentVNodePath()}}},
						&components.LabelInline{Title: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$in")}}},
						&components.LabelInline{Title: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$in")}}},
						&components.LabelInline{Title: "Items", Children: []components.PageInterface{&components.FieldText{Getter: vnodeChildrenCountForKey("$in")}}},
						&components.LabelInline{Title: "Created At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.CreatedAt")}}},
						&components.LabelInline{Title: "Modified At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.UpdatedAt")}}},
						&components.ShowIf{
							Getter: getters.Any(func(ctx context.Context) (bool, error) {
								isDirectory, err := getters.Key[bool]("$in.IsDirectory")(ctx)
								if err != nil {
									return false, err
								}
								return !isDirectory, nil
							}),
							Children: []components.PageInterface{
								&components.ButtonDownload{Label: "Download", Link: lago.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
									"id": getters.Any(getters.Key[uint]("$in.ID")),
								}), Icon: "arrow-down-tray", Classes: "btn-primary mt-4"},
							},
						},
					}},
				},
			},
		},
	})
}

func selectionTable(name, filterName, childRoute string, multi, selectDirectories bool) *components.Modal {
	title := "Select Directory"
	subtitle := "Choose a folder"
	if !selectDirectories {
		title = "Select Files"
		subtitle = "Choose files, or open folders to browse deeper"
	}

	modalID := "filesystem-selection-modal-" + name
	onClick := selectionRowClickGetter(name, modalID, childRoute, multi, selectDirectories)
	var rowClass getters.Getter[string]
	if multi {
		rowClass = getters.SelectMultiRowClass(selectionTargetInput(name), getters.Key[uint]("$row.ID"))
	}

	return &components.Modal{
		UID: modalID,
		Children: []components.PageInterface{
			&components.DataTable[VNode]{
				UID:      "filesystem-selection-table-" + name,
				Title:    title,
				Subtitle: subtitle,
				Data:     getters.Key[components.ObjectList[VNode]]("vnodes"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: filterName}},
				},
				RowAttr: getters.RowAttrClickWithClass(onClick, rowClass),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Path", Name: "Path", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")}}},
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
	lago.RegistryPage.Register("filesystem.VNodeDeleteForm", &components.Modal{
		UID: "filesystem-vnode-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this item? Deleting directories will remove all nested contents.",
				Attr:    getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(currentVNodeDeleteRoute())),
			},
		},
	})
}
