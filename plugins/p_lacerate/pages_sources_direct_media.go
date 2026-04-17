package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func directMediaSourceListTableConfig() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.DirectMediaSourceListCell"},
		Children: []components.PageInterface{
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$row.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					return strings.HasPrefix(strings.TrimSpace(s), directMediaFSURLPrefix), nil
				})),
				Children: []components.PageInterface{
					&p_filesystem.FieldFile{
						VNode: getters.Association[p_filesystem.VNode](
							getters.Map(getters.Key[string]("$row.DirectMedia.URL"), func(_ context.Context, s string) (uint, error) {
								return parseDirectMediaFSURLNodeID(s)
							}),
						),
					},
				},
			},
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$row.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					s = strings.TrimSpace(s)
					return s != "" && !strings.HasPrefix(s, directMediaFSURLPrefix), nil
				})),
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.DirectMedia.URL")},
				},
			},
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$row.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					return strings.TrimSpace(s) == "", nil
				})),
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Static("(none)")},
				},
			},
		},
	}
}

func directMediaSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.DirectMediaSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.URL"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Direct media URL",
						Name:    "URL",
						Getter:  getters.Key[string]("$in.DirectMedia.URL"),
						Classes: "w-full",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.DirectMediaUpload"),
				Children: []components.PageInterface{
					&components.InputFile{
						Label:   "Or upload a file",
						Name:    "DirectMediaUpload",
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func directMediaSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailDirectMediaFields"},
		Children: []components.PageInterface{
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$in.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					return strings.HasPrefix(strings.TrimSpace(s), directMediaFSURLPrefix), nil
				})),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Upload",
						Children: []components.PageInterface{
							&p_filesystem.FieldFile{
								VNode: getters.Association[p_filesystem.VNode](
									getters.Map(getters.Key[string]("$in.DirectMedia.URL"), func(_ context.Context, s string) (uint, error) {
										return parseDirectMediaFSURLNodeID(s)
									}),
								),
							},
						},
					},
				},
			},
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$in.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					s = strings.TrimSpace(s)
					return s != "" && !strings.HasPrefix(s, directMediaFSURLPrefix), nil
				})),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "URL",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$in.DirectMedia.URL")},
						},
					},
				},
			},
			&components.ShowIf{
				Getter: getters.Any(getters.Map(getters.Key[string]("$in.DirectMedia.URL"), func(_ context.Context, s string) (bool, error) {
					return strings.TrimSpace(s) == "", nil
				})),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "URL",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Static("(none)")},
						},
					},
				},
			},
		},
	}
}
