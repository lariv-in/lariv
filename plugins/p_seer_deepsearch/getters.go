package p_seer_deepsearch

import (
	"context"
	"sort"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

func deepSearchPollingActive(status string) bool {
	switch status {
	case DeepSearchStatusDone, DeepSearchStatusFailed, DeepSearchStatusCancelled:
		return false
	default:
		return true
	}
}

// deepSearchHomeFormAttr merges [getters.FormBubbling] with an explicit form action so
// [components.FormListenBoostedPost] path checks pass (form pathname must match POST target).
func deepSearchHomeFormAttr() getters.Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		actionURL, err := lago.RoutePath("seer_deepsearch.StartRoute", nil)(ctx)
		if err != nil {
			return nil, err
		}
		bub, err := getters.FormBubbling(deepSearchHomeFormName)(ctx)
		if err != nil {
			return nil, err
		}
		return gomponents.Group{
			ghtml.Action(actionURL),
			bub,
		}, nil
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
