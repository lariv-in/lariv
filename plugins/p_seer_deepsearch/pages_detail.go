package p_seer_deepsearch

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/registry"
)

func deepSearchPollingActive(status string) bool {
	switch status {
	case DeepSearchStatusDone, DeepSearchStatusFailed, DeepSearchStatusCancelled:
		return false
	default:
		return true
	}
}

func deepSearchShowStopActionGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		ds, err := getters.Key[DeepSearch]("deepSearch")(ctx)
		if err != nil {
			return false, err
		}
		switch ds.Status {
		case DeepSearchStatusPending, DeepSearchStatusRunning, DeepSearchStatusExpandingQueries,
			DeepSearchStatusSearching, DeepSearchStatusScraping, DeepSearchStatusIngestingIntel,
			DeepSearchStatusReporting:
			return true, nil
		default:
			return false, nil
		}
	}
}

func deepSearchShowRestartActionGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		ds, err := getters.Key[DeepSearch]("deepSearch")(ctx)
		if err != nil {
			return false, err
		}
		switch ds.Status {
		case DeepSearchStatusDone, DeepSearchStatusFailed, DeepSearchStatusCancelled:
			return true, nil
		default:
			return false, nil
		}
	}
}

func deepSearchDetailPollURL(ctx context.Context, id uint) (string, error) {
	return lago.RoutePath("seer_deepsearch.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(id), 10))),
	})(ctx)
}

func deepSearchDetailShellGetter(inner components.PageInterface) getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		ds, err := getters.Key[DeepSearch]("deepSearch")(ctx)
		if err != nil {
			return nil, err
		}
		if !deepSearchPollingActive(ds.Status) {
			return inner, nil
		}
		u, err := deepSearchDetailPollURL(ctx, ds.ID)
		if err != nil {
			return nil, err
		}
		return &components.HTMXPolling{
			Page:     components.Page{Key: "seer_deepsearch.DeepSearchDetailPolling"},
			URL:      getters.Static(u),
			Children: []components.PageInterface{inner},
		}, nil
	}
}

func deepSearchLogsGetter() getters.Getter[[]DeepSearchLog] {
	return func(ctx context.Context) ([]DeepSearchLog, error) {
		ds, err := getters.Key[DeepSearch]("deepSearch")(ctx)
		if err != nil {
			return nil, err
		}
		if len(ds.Logs) == 0 {
			return []DeepSearchLog{}, nil
		}
		out := append([]DeepSearchLog(nil), ds.Logs...)
		sort.Slice(out, func(i, j int) bool {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		})
		return out, nil
	}
}

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
