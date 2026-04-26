package p_seer_deepsearch

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/registry"
)

func deepSearchDetailContentColumn() components.PageInterface {
	return components.ContainerColumn{
		Page: components.Page{Key: "seer_deepsearch.DeepSearchDetailContent"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_deepsearch.DeepSearchDetailActions"},
				Classes: "flex flex-wrap gap-2 mb-4",
				Children: []components.PageInterface{
					&components.ShowIf{
						Page:   components.Page{Key: "seer_deepsearch.DeepSearchDetailStopWrap"},
						Getter: deepSearchShowStopActionGetter(),
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_deepsearch.DeepSearchDetailStopBtn"},
								Label:   "Stop pipeline",
								URL:     lago.RoutePath("seer_deepsearch.StopRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("deepSearch.ID"))}),
								Icon:    "stop",
								Classes: "btn-outline btn-warning btn-sm",
							},
						},
					},
					&components.ShowIf{
						Page:   components.Page{Key: "seer_deepsearch.DeepSearchDetailRestartWrap"},
						Getter: deepSearchShowRestartActionGetter(),
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_deepsearch.DeepSearchDetailRestartBtn"},
								Label:   "Restart pipeline",
								URL:     lago.RoutePath("seer_deepsearch.RestartRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("deepSearch.ID"))}),
								Icon:    "arrow-path",
								Classes: "btn-outline btn-primary btn-sm",
							},
						},
					},
				},
			},
			&components.LabelInline{
				Title: "Question",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter:  getters.Key[string]("$in.Query"),
						Classes: "whitespace-pre-wrap",
					},
				},
			},
			&components.LabelInline{
				Title: "Status",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: registry.PairValueFromKey(getters.Key[string]("$in.Status"), DeepSearchStatusChoices),
					},
				},
			},
			&components.ShowIf{
				Page: components.Page{Key: "seer_deepsearch.DeepSearchDetailErrorWrap"},
				Getter: func(ctx context.Context) (any, error) {
					ds, err := getters.Key[DeepSearch]("deepSearch")(ctx)
					if err != nil {
						return false, err
					}
					return ds.RunError != "", nil
				},
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Error",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter:  getters.Key[string]("$in.RunError"),
								Classes: "whitespace-pre-wrap text-error",
							},
						},
					},
				},
			},
			&components.LabelInline{
				Title: "Activity log",
				Children: []components.PageInterface{
					&components.FieldList[DeepSearchLog]{
						Page:    components.Page{Key: "seer_deepsearch.DeepSearchLogList"},
						Getter:  deepSearchLogsGetter(),
						Classes: "border border-base-300 rounded-md p-2 max-h-[32rem] overflow-y-auto text-sm",
						Children: []components.PageInterface{
							&components.ContainerRow{
								Page:    components.Page{Key: "seer_deepsearch.DeepSearchLogRow"},
								Classes: "flex flex-wrap gap-x-4 gap-y-1 border-b border-base-200 py-1.5 last:border-0",
								Children: []components.PageInterface{
									&components.FieldDatetime{
										Getter:  getters.Key[time.Time]("$row.CreatedAt"),
										Classes: "shrink-0 text-xs opacity-80",
									},
									&components.FieldText{
										Getter:  registry.PairValueFromKey(getters.Key[string]("$row.Kind"), DeepSearchLogKindChoices),
										Classes: "font-medium shrink-0 w-40",
									},
									&components.FieldText{
										Getter:  getters.Key[string]("$row.Message"),
										Classes: "whitespace-pre-wrap flex-1 min-w-0",
									},
								},
							},
						},
					},
				},
			},
			&components.LabelInline{
				Title: "Report",
				Children: []components.PageInterface{
					&components.FieldMarkdown{
						Getter:      getters.Key[string]("$in.Report"),
						Classes:     "prose max-w-none",
						RenderHooks: p_seer_intel.IntelProtocolMarkdownHooks,
					},
				},
			},
		},
	}
}

func registerDeepSearchDetailPages() {
	inner := deepSearchDetailContentColumn()
	lago.RegistryPage.Register("seer_deepsearch.DeepSearchDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_deepsearch.DeepSearchMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[DeepSearch]{
				Getter: getters.Key[DeepSearch]("deepSearch"),
				Children: []components.PageInterface{
					&components.GetterPage{
						Page:   components.Page{Key: "seer_deepsearch.DeepSearchDetailShell"},
						Getter: deepSearchDetailShellGetter(inner),
					},
				},
			},
		},
	})
}
