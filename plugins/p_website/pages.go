package p_website

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
)

func formatTimeVal(val any, fmtLayout string) string {
	if val == nil {
		return ""
	}
	var t time.Time
	switch v := val.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v == nil {
			return ""
		}
		t = *v
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return ""
		}
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339Nano, v)
		}
		if err != nil {
			parsed, err = time.Parse("2006-01-02 15:04:05", v)
		}
		if err != nil {
			parsed, err = time.Parse("2006-01-02", v)
		}
		if err != nil {
			return v
		}
		t = parsed
	default:
		return fmt.Sprintf("%v", val)
	}
	return t.Format(fmtLayout)
}

// DynamicWebsitePage renders database-driven website pages or streams static files.
//
// Note: We cannot allow a dynamic view that will serve any arbitrary file under a directory,
// since it might give arbitrary read access using Go templates and the custom filesystem.
type DynamicWebsitePage struct {
	components.Page
}

func (p DynamicWebsitePage) GetKey() string {
	return p.Key
}

func (p DynamicWebsitePage) GetRoles() []string {
	return p.Roles
}

func pathToLTree(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "root"
	}
	r := strings.ReplaceAll(trimmed, "/", ".")
	r = strings.ReplaceAll(r, "-", "_")
	return r
}

func pathToLQuery(pattern string) string {
	trimmed := strings.Trim(pattern, "/")
	if trimmed == "" {
		return "root"
	}
	r := strings.ReplaceAll(trimmed, "/", ".")
	r = strings.ReplaceAll(r, "-", "_")
	return r
}

func matchGoWildcard(pattern, reqPath string) bool {
	pattern = strings.Trim(pattern, "/")
	reqPath = strings.Trim(reqPath, "/")

	if pattern == reqPath || pattern == "*" {
		return true
	}

	pParts := strings.Split(pattern, "/")
	rParts := strings.Split(reqPath, "/")

	if len(pParts) != len(rParts) {
		if len(pParts) > 0 && (pParts[len(pParts)-1] == "*" || strings.HasPrefix(pParts[len(pParts)-1], "*{")) {
			if len(rParts) >= len(pParts)-1 {
				prefixMatched := true
				for i := 0; i < len(pParts)-1; i++ {
					if pParts[i] != "*" && pParts[i] != rParts[i] {
						prefixMatched = false
						break
					}
				}
				if prefixMatched {
					return true
				}
			}
		}
		return false
	}

	for i := 0; i < len(pParts); i++ {
		if pParts[i] != "*" && pParts[i] != rParts[i] {
			return false
		}
	}
	return true
}

func FindMatchingDBRoute(db *gorm.DB, reqPath string) (*DBRoute, error) {
	reqLTree := pathToLTree(reqPath)

	// 1. Exact match on path
	var route DBRoute
	err := db.Preload("Page").Preload("References").
		Where("path = ? AND is_active = ?", reqPath, true).
		First(&route).Error
	if err == nil {
		return &route, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Exact match on ltree_path
	err = db.Preload("Page").Preload("References").
		Where("ltree_path = ?::ltree AND is_active = ?", reqLTree, true).
		First(&route).Error
	if err == nil {
		return &route, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 3. Evaluate lquery / ltxtquery / wildcard matches
	var activeRoutes []DBRoute
	if err := db.Preload("Page").Preload("References").
		Where("is_active = ?", true).
		Find(&activeRoutes).Error; err != nil {
		return nil, err
	}

	for _, r := range activeRoutes {
		lqueryPattern := pathToLQuery(r.Path)
		var matched bool

		if db.Dialector.Name() == "postgres" {
			var count int64
			// Check lquery match: reqLTree ~ lqueryPattern
			err := db.Raw("SELECT COUNT(*) FROM (SELECT ?::ltree AS t) sub WHERE t ~ ?::lquery", reqLTree, lqueryPattern).Scan(&count).Error
			if err == nil && count > 0 {
				matched = true
			}

			if !matched {
				// Check reverse lquery match: r.LTreePath ~ reqLTree::lquery
				err = db.Raw("SELECT COUNT(*) FROM (SELECT ?::ltree AS t) sub WHERE t ~ ?::lquery", r.LTreePath, reqLTree).Scan(&count).Error
				if err == nil && count > 0 {
					matched = true
				}
			}
		}

		if !matched {
			matched = matchGoWildcard(r.Path, reqPath)
		}

		if matched {
			return &r, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func (p DynamicWebsitePage) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("DynamicWebsitePage: failed to resolve DB from context", "error", err)
		return Text("Internal Server Error")
	}

	req, ok := ctx.Value("$request").(*http.Request)
	if !ok || req == nil {
		slog.Error("DynamicWebsitePage: missing or nil $request in context")
		return Text("Internal Server Error")
	}

	dbRoutePtr, err := FindMatchingDBRoute(db, req.URL.Path)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Text("404 Not Found")
		}
		slog.Error("DynamicWebsitePage: failed to query DBRoute", "path", req.URL.Path, "error", err)
		return Text("Internal Server Error")
	}
	dbRoute := *dbRoutePtr

	vnodePath := dbRoute.Page.GetPath(db)
	relPath := strings.TrimPrefix(vnodePath, "/")

	dbFS := p_filesystem.NewDatabaseFilesystem(db)

	ext := strings.ToLower(filepath.Ext(relPath))
	if slices.Contains([]string{".html", ".tmpl", ".htm", ".js", ".css", ".txt", ".md", ".json", ".yaml", ".yml"}, ext) {
		funcMap := template.FuncMap{
			"query": func(tableName string, limit, offset int) ([]map[string]any, error) {
				if limit <= 0 {
					limit = 10
				}
				var results []map[string]any
				err := db.Table(tableName).Limit(limit).Offset(offset).Find(&results).Error
				return results, err
			},
			"query_where": func(tableName string, query string, args ...any) ([]map[string]any, error) {
				var results []map[string]any
				err := db.Table(tableName).Where(query, args...).Find(&results).Error
				return results, err
			},
			"m2m_list": func(leftTable, m2mTable, rightTable string, id any) ([]map[string]any, error) {
				leftCol := strings.TrimSuffix(leftTable, "s") + "_id"
				rightCol := strings.TrimSuffix(rightTable, "s") + "_id"
				var results []map[string]any
				err := db.Table(rightTable).
					Joins(fmt.Sprintf("JOIN %s ON %s.%s = %s.id", m2mTable, m2mTable, rightCol, rightTable)).
					Where(fmt.Sprintf("%s.%s = ?", m2mTable, leftCol), id).
					Find(&results).Error
				return results, err
			},
			"m2o": func(leftTable, rightTable string, id any) (map[string]any, error) {
				var leftRows []map[string]any
				if err := db.Table(leftTable).Where("id = ?", id).Limit(1).Find(&leftRows).Error; err != nil || len(leftRows) == 0 {
					return nil, err
				}
				leftRow := leftRows[0]
				fkCol := strings.TrimSuffix(rightTable, "s") + "_id"
				fkVal, ok := leftRow[fkCol]
				if !ok || fkVal == nil {
					if val, exists := leftRow["created_by_id"]; exists {
						fkVal = val
						ok = true
					}
				}
				if !ok || fkVal == nil {
					return nil, nil
				}
				var rightRows []map[string]any
				if err := db.Table(rightTable).Where("id = ?", fkVal).Limit(1).Find(&rightRows).Error; err != nil || len(rightRows) == 0 {
					return nil, err
				}
				return rightRows[0], nil
			},
			"get": func(tableName string, id any) (map[string]any, error) {
				var results []map[string]any
				err := db.Table(tableName).Where("id = ?", id).Limit(1).Find(&results).Error
				if err != nil || len(results) == 0 {
					return nil, err
				}
				return results[0], nil
			},
			"format_datetime": func(val any, layout ...string) string {
				fmtLayout := "Mon, 02 Jan 2006 15:04:05"
				if len(layout) > 0 && layout[0] != "" {
					fmtLayout = layout[0]
				}
				return formatTimeVal(val, fmtLayout)
			},
			"format_date": func(val any, layout ...string) string {
				fmtLayout := "02 Jan 2006"
				if len(layout) > 0 && layout[0] != "" {
					fmtLayout = layout[0]
				}
				return formatTimeVal(val, fmtLayout)
			},
			"slug": func() string {
				if req == nil || req.URL == nil {
					return ""
				}
				return filepath.Base(strings.TrimSuffix(req.URL.Path, "/"))
			},
			"path": func() string {
				if req == nil || req.URL == nil {
					return ""
				}
				return req.URL.Path
			},
			"param": func(name string) string {
				if req == nil || req.URL == nil {
					return ""
				}
				return req.URL.Query().Get(name)
			},
			"first": func(slice []map[string]any) map[string]any {
				if len(slice) == 0 {
					return nil
				}
				return slice[0]
			},
			"markdown": func(val any) template.HTML {
				if val == nil {
					return ""
				}
				s, ok := val.(string)
				if !ok {
					s = fmt.Sprintf("%v", val)
				}
				return template.HTML(components.RenderMarkdown(s))
			},
		}

		patterns := []string{relPath}
		for _, ref := range dbRoute.References {
			refPath := ref.GetPath(db)
			refRelPath := strings.TrimPrefix(refPath, "/")
			if refRelPath != "" && !slices.Contains(patterns, refRelPath) {
				patterns = append(patterns, refRelPath)
			}
		}

		tmplComp := &components.TemplateFSComponent{
			Page: components.Page{
				Key: fmt.Sprintf("db_template_%d", dbRoute.Page.ID),
			},
			Filesystem:       dbFS,
			TemplatePatterns: patterns,
			TemplateName:     "",
			TemplateContext: getters.Getter[any](func(ctx context.Context) (any, error) {
				return ctx, nil
			}),
			Funcs: funcMap,
		}
		return components.Render(tmplComp, ctx)
	}

	return fileStreamingNode{
		dbFS:    dbFS,
		relPath: relPath,
	}
}

type fileStreamingNode struct {
	dbFS    lariv.UsefulFilesystem
	relPath string
}

func (f fileStreamingNode) Render(w io.Writer) error {
	file, err := f.dbFS.Open(f.relPath)
	if err != nil {
		slog.Error("fileStreamingNode: failed to open file", "path", f.relPath, "error", err)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, "404 Not Found", http.StatusNotFound)
			return nil
		}
		_, err = io.WriteString(w, "404 Not Found")
		return err
	}
	defer file.Close()

	if rw, ok := w.(http.ResponseWriter); ok {
		ext := strings.ToLower(filepath.Ext(f.relPath))
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		rw.Header().Set("Content-Type", contentType)
	}

	_, err = io.Copy(w, file)
	return err
}

func routeFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Path"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Path", Name: "Path", Required: true, Getter: getters.Key[string]("$in.Path")},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.PageID"),
				Children: []components.PageInterface{
					&p_filesystem.InputFile{
						Label:    "Template Page",
						Name:     "PageID",
						Required: true,
						VNode:    getters.Key[p_filesystem.VNode]("$in.Page"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.References"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_filesystem.VNode]{
						Label:       "Reference Files",
						Name:        "References",
						Url:         lariv.RoutePath("filesystem.MultiSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select reference files...",
						Required:    false,
						Getter:      getters.Key[[]p_filesystem.VNode]("$in.References"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.IsActive"),
				Children: []components.PageInterface{
					&components.InputCheckbox{Label: "Is Active", Name: "IsActive", Getter: getters.Key[bool]("$in.IsActive")},
				},
			},
		},
	}
}

func pluginPages() lariv.PluginFeatures[components.PageInterface] {
	return lariv.PluginFeatures[components.PageInterface]{
		Entries: []registry.Pair[string, components.PageInterface]{
			{Key: "p_website.DynamicWebsitePage", Value: DynamicWebsitePage{}},
			{Key: "p_website.RoutesListMenu", Value: &components.SidebarMenu{
				Title: getters.Static("Website Admin"),
				Back: &components.SidebarMenuItem{
					Title: getters.Static("Back to All Apps"),
					Url:   lariv.RoutePath("dashboard.AppsPage", nil),
				},
				Children: []components.PageInterface{
					&components.SidebarMenuItem{
						Title: getters.Static("All Routes"),
						Url:   lariv.RoutePath("p_website.RoutesListRoute", nil),
					},
				},
			}},
			{Key: "p_website.RoutesDetailMenu", Value: &components.SidebarMenu{
				Title: getters.Format("Route: %s", getters.Any(getters.Key[string]("dbroute.Path"))),
				Back: &components.SidebarMenuItem{
					Title: getters.Static("Back to All Routes"),
					Url:   lariv.RoutePath("p_website.RoutesListRoute", nil),
				},
				Children: []components.PageInterface{
					&components.SidebarMenuItem{
						Title: getters.Static("Route Details"),
						Url: lariv.RoutePath("p_website.RoutesDetailRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("dbroute.ID")),
						}),
					},
					&components.SidebarMenuItem{
						Title: getters.Static("Edit Route"),
						Url: lariv.RoutePath("p_website.RoutesUpdateRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("dbroute.ID")),
						}),
					},
				},
			}},
			{Key: "p_website.RoutesListPage", Value: &components.ShellScaffold{
				Sidebar: []components.PageInterface{
					lariv.DynamicPage{Name: "p_website.RoutesListMenu"},
				},
				Children: []components.PageInterface{
					&components.DataTable[DBRoute]{
						UID:     "routes-table",
						Classes: "w-full",
						Data:    getters.Key[components.ObjectList[DBRoute]]("dbroutes"),
						Actions: []components.PageInterface{
							&components.ButtonLink{
								Link:    lariv.RoutePath("p_website.RoutesCreateRoute", nil),
								Icon:    "plus",
								Classes: "btn-square btn-outline btn-sm",
							},
						},
						RowAttr: getters.RowAttrNavigate(lariv.RoutePath("p_website.RoutesDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
						Columns: []components.TableColumn{
							{Label: "Path", Name: "Path", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$row.Path")},
							}},
							{Label: "Template Node", Name: "Page", Children: []components.PageInterface{
								&components.FieldText{Getter: getters.Key[string]("$row.Page.Name")},
							}},
							{Label: "Is Active", Name: "IsActive", Children: []components.PageInterface{
								&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsActive")},
							}},
						},
					},
				},
			}},
			{Key: "p_website.RoutesDetailPage", Value: &components.ShellScaffold{
				Sidebar: []components.PageInterface{
					lariv.DynamicPage{Name: "p_website.RoutesDetailMenu"},
				},
				Children: []components.PageInterface{
					&components.Detail[DBRoute]{
						Page:   components.Page{Key: "p_website.RoutesDetailContent"},
						Getter: getters.Key[DBRoute]("dbroute"),
						Children: []components.PageInterface{
							&components.ContainerColumn{
								Children: []components.PageInterface{
									&components.FieldTitle{Getter: getters.Key[string]("$in.Path")},
									&components.FieldLink{
										Href:    getters.Key[string]("$in.Path"),
										Label:   getters.Static("View Live Page ↗"),
										Classes: "link link-primary font-semibold mb-4 block",
										Attr:    getters.Static[Node](Attr("hx-boost", "false")),
									},
									&components.LabelInline{
										Title:   "Template Page Name",
										Classes: "mt-4 block",
										Children: []components.PageInterface{
											&p_filesystem.FieldFile{VNode: getters.Key[p_filesystem.VNode]("$in.Page")},
										},
									},
									&components.LabelInline{
										Title:   "Reference Files",
										Classes: "mt-4 block",
										Children: []components.PageInterface{
											&p_filesystem.FieldManyFile{VNode: getters.Key[[]p_filesystem.VNode]("$in.References")},
										},
									},
									&components.LabelInline{
										Title:   "Is Active",
										Classes: "mt-4 block",
										Children: []components.PageInterface{
											&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsActive")},
										},
									},
								},
							},
						},
					},
				},
			}},
			{Key: "p_website.RoutesCreatePage", Value: &components.ShellScaffold{
				Sidebar: []components.PageInterface{
					lariv.DynamicPage{Name: "p_website.RoutesListMenu"},
				},
				Children: []components.PageInterface{
					&components.FormListenBoostedPost{
						Name:      getters.Static("p_website.RoutesCreateForm"),
						ActionURL: lariv.RoutePath("p_website.RoutesCreateRoute", nil),
						Children: []components.PageInterface{
							&components.FormComponent[DBRoute]{
								Page: components.Page{Key: "p_website.RoutesCreateForm"},
								Attr: getters.FormBubbling(getters.Static("p_website.RoutesCreateForm")),
								ChildrenInput: []components.PageInterface{
									routeFormFields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ButtonSubmit{Label: "Create Route"},
								},
							},
						},
					},
				},
			}},
			{Key: "p_website.RoutesUpdatePage", Value: &components.ShellScaffold{
				Sidebar: []components.PageInterface{
					lariv.DynamicPage{Name: "p_website.RoutesDetailMenu"},
				},
				Children: []components.PageInterface{
					&components.FormListenBoostedPost{
						Name: getters.Static("p_website.RoutesUpdateForm"),
						ActionURL: lariv.RoutePath("p_website.RoutesUpdateRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("dbroute.ID")),
						}),
						Children: []components.PageInterface{
							&components.FormComponent[DBRoute]{
								Page:   components.Page{Key: "p_website.RoutesUpdateForm"},
								Attr:   getters.FormBubbling(getters.Static("p_website.RoutesUpdateForm")),
								Getter: getters.Key[DBRoute]("dbroute"),
								ChildrenInput: []components.PageInterface{
									routeFormFields(),
								},
								ChildrenAction: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
										Children: []components.PageInterface{
											&components.ContainerRow{
												Classes: "flex justify-end gap-2",
												Children: []components.PageInterface{
													&components.ButtonSubmit{Label: "Save Changes"},
													&components.ButtonModalForm{
														Label:       "Delete",
														Icon:        "trash",
														Name:        getters.Static("p_website.RoutesDeleteForm"),
														Url:         lariv.RoutePath("p_website.RoutesDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("dbroute.ID"))}),
														FormPostURL: lariv.RoutePath("p_website.RoutesDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("dbroute.ID"))}),
														ModalUID:    "routes-delete-modal",
														Classes:     "btn-error",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}},
			{Key: "p_website.RoutesDeleteForm", Value: &components.Modal{
				UID: "routes-delete-modal",
				Children: []components.PageInterface{
					&components.DeleteConfirmation{
						Page:    components.Page{Key: "p_website.RoutesDeleteForm"},
						Title:   "Confirm deletion",
						Message: "Are you sure you want to delete this route? This action cannot be undone.",
						Attr:    getters.FormBubbling(getters.Static("p_website.RoutesDeleteForm")),
					},
				},
			}},
		},
	}
}
