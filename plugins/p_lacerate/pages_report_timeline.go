package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func timelineEntriesJSONValue(entries []TimelineReportEntry, tz *time.Location) string {
	if len(entries) == 0 {
		return "[]"
	}
	if tz == nil {
		tz = time.UTC
	}
	out := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, map[string]string{
			"datetime": entry.Datetime.In(tz).Format("2006-01-02T15:04"),
			"title":    entry.Title,
			"content":  entry.Content,
		})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(b)
}

func timelineEntriesJSONGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if s, err := getters.Key[string]("$in.TimelineEntriesJSON")(ctx); err == nil {
			return s, nil
		}
		timeline, err := getters.Key[*TimelineReport]("$in.Timeline")(ctx)
		if err != nil || timeline == nil || len(timeline.Entries) == 0 {
			return "[]", nil
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		if tz == nil {
			tz = time.UTC
		}
		return timelineEntriesJSONValue(timeline.Entries, tz), nil
	}
}

func timelineReportListTableConfig() components.PageInterface {
	return &components.FieldText{Getter: getters.IfOrElse(
		getters.Map(getters.Key[ReportPageData]("$row"), func(_ context.Context, data ReportPageData) (string, error) {
			if data.Timeline == nil || len(data.Timeline.Entries) == 0 {
				return "", nil
			}
			first := data.Timeline.Entries[0]
			last := data.Timeline.Entries[len(data.Timeline.Entries)-1]
			if len(data.Timeline.Entries) == 1 {
				return fmt.Sprintf("1 entry · %s", first.Datetime.Format("2006-01-02 15:04")), nil
			}
			return fmt.Sprintf("%d entries · %s -> %s", len(data.Timeline.Entries), first.Datetime.Format("2006-01-02 15:04"), last.Datetime.Format("2006-01-02 15:04")), nil
		}),
		getters.Static("—"),
	)}
}

func timelineReportFormFields() components.PageInterface {
	return &components.ContainerError{
		Error: getters.Key[error]("$error.TimelineEntriesJSON"),
		Children: []components.PageInterface{
			&InputTimelineEntries{
				Label:   "Timeline entries",
				Name:    "TimelineEntriesJSON",
				Classes: "w-full",
				Getter:  timelineEntriesJSONGetter(),
			},
		},
	}
}

func timelineReportDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.ReportTimelineDetailFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "Entries",
				Children: []components.PageInterface{
					&components.FieldList[TimelineReportEntry]{
						Page:    components.Page{Key: "lacerate.ReportTimelineEntries"},
						Getter: func(ctx context.Context) ([]TimelineReportEntry, error) {
							if timeline, err := getters.Key[*TimelineReport]("$in.Timeline")(ctx); err == nil && timeline != nil {
								return timeline.Entries, nil
							}
							return nil, nil
						},
						Classes: "space-y-4",
						Children: []components.PageInterface{
							&components.ContainerColumn{
								Classes: "border border-base-300 rounded p-3 gap-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Map(getters.Key[string]("$row.Title"), func(_ context.Context, s string) (string, error) {
										s = strings.TrimSpace(s)
										if s == "" {
											return "(untitled)", nil
										}
										return s, nil
									})},
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime"), Classes: "text-sm opacity-80"},
									&components.FieldMarkdown{Getter: getters.Key[string]("$row.Content"), Classes: "prose prose-sm max-w-none"},
								},
							},
						},
					},
				},
			},
		},
	}
}
