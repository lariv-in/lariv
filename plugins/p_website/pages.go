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

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/gorm"
	. "maragu.dev/gomponents"
)

type DynamicWebsitePage struct {
	components.Page
}

func (p DynamicWebsitePage) GetKey() string {
	return p.Key
}

func (p DynamicWebsitePage) GetRoles() []string {
	return p.Roles
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

	var dbRoute DBRoute
	err = db.Preload("Page").Where("path = ? AND is_active = ?", req.URL.Path, true).First(&dbRoute).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Text("404 Not Found")
		}
		slog.Error("DynamicWebsitePage: failed to query DBRoute", "path", req.URL.Path, "error", err)
		return Text("Internal Server Error")
	}

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
			"get": func(tableName string, id any) (map[string]any, error) {
				var result map[string]any
				err := db.Table(tableName).Where("id = ?", id).First(&result).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, nil
					}
					return nil, err
				}
				return result, nil
			},
		}

		tmplComp := &components.TemplateFSComponent{
			Page: components.Page{
				Key: fmt.Sprintf("db_template_%d", dbRoute.Page.ID),
			},
			Filesystem:       dbFS,
			TemplatePatterns: []string{relPath},
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
