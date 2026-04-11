package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func briefingReportContentGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if s, err := getters.Key[string]("$in.BriefingContent")(ctx); err == nil {
			return s, nil
		}
		if s, err := getters.Key[string]("$in.Briefing.Content")(ctx); err == nil {
			return s, nil
		}
		return "", nil
	}
}

func briefingReportListTableConfig() components.PageInterface {
	return &components.FieldText{Getter: getters.IfOrElse(
		getters.Map(briefingReportSnippetGetter("$row"), func(_ context.Context, s string) (string, error) {
			s = strings.TrimSpace(s)
			if s == "" {
				return "", nil
			}
			if len(s) > 96 {
				return s[:93] + "...", nil
			}
			return s, nil
		}),
		getters.Static("—"),
	)}
}

func briefingReportFormFields() components.PageInterface {
	return &components.ContainerError{
		Error: getters.Key[error]("$error.BriefingContent"),
		Children: []components.PageInterface{
			&components.InputTextarea{
				Label:    "Briefing content",
				Name:     "BriefingContent",
				Required: true,
				Rows:     18,
				Classes:  "w-full font-mono text-sm",
				Getter:   briefingReportContentGetter(),
			},
		},
	}
}

func briefingReportDetailFields() components.PageInterface {
	return &components.LabelInline{
		Title: "Briefing",
		Children: []components.PageInterface{
			&components.FieldMarkdown{
				Getter:  getters.Key[string]("$in.Briefing.Content"),
				Classes: "prose prose-sm max-w-none",
			},
		},
	}
}

func briefingReportSnippetGetter(path string) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		data, err := getters.Key[ReportPageData](path)(ctx)
		if err != nil {
			return "", err
		}
		return reportPageDataSnippet(data), nil
	}
}
