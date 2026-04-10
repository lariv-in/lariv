package p_lacerate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
)

func lookupTouchedTargetsOfInterestListGetter() getters.Getter[[]LookupTouchedTargetOfInterestDisplay] {
	return func(ctx context.Context) ([]LookupTouchedTargetOfInterestDisplay, error) {
		raw := ctx.Value(ctxKeyLookupTouchedTargetsOfInterest)
		if raw == nil {
			return nil, nil
		}
		v, ok := raw.([]LookupTouchedTargetOfInterestDisplay)
		if !ok {
			return nil, nil
		}
		return v, nil
	}
}

func touchedTargetOfInterestRowLabelGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		row, ok := ctx.Value("$row").(LookupTouchedTargetOfInterestDisplay)
		if !ok {
			return "", nil
		}
		if row.TargetOfInterest.ID == 0 {
			return "(missing Target of Interest)", nil
		}
		name := strings.TrimSpace(row.TargetOfInterest.Name)
		if name == "" {
			name = fmt.Sprintf("#%d", row.TargetOfInterest.ID)
		}
		actLabel := row.Action
		if p, ok := registry.PairFromPairs(row.Action, LookupTargetOfInterestTouchActionChoices); ok {
			actLabel = p.Value
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		if tz == nil {
			tz = time.UTC
		}
		ts := row.LogCreatedAt.In(tz).Format(time.RFC3339)
		return fmt.Sprintf("%s — %s · %s", name, actLabel, ts), nil
	}
}

func touchedTargetOfInterestRowLinkGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		row, ok := ctx.Value("$row").(LookupTouchedTargetOfInterestDisplay)
		if !ok || row.TargetOfInterest.ID == 0 {
			return "", nil
		}
		g := lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(row.TargetOfInterest.ID)),
		})
		return g(ctx)
	}
}

func lookupDetailTouchedTargetsOfInterestSection() components.PageInterface {
	return &components.ContainerColumn{
		Page:    components.Page{Key: "lacerate.LookupDetailTouchedTargetsOfInterest"},
		Classes: "w-full mt-8",
		Children: []components.PageInterface{
			&components.FieldTitle{
				Getter:  getters.Static("Targets of Interest touched by agent"),
				Classes: "mb-3",
			},
			&components.FieldList[LookupTouchedTargetOfInterestDisplay]{
				Page:    components.Page{Key: "lacerate.LookupDetailTouchedTargetsOfInterestList"},
				Getter:  lookupTouchedTargetsOfInterestListGetter(),
				Classes: "space-y-2",
				Children: []components.PageInterface{
					&components.ButtonLink{
						GetterLabel: touchedTargetOfInterestRowLabelGetter(),
						Link:        touchedTargetOfInterestRowLinkGetter(),
						Icon:        "document-text",
						Classes:     "btn btn-ghost btn-sm justify-start h-auto min-h-10 whitespace-normal text-left",
					},
				},
			},
		},
	}
}

func lookupLogEntriesObjectListGetter() getters.Getter[components.ObjectList[LookupLogDisplay]] {
	return func(ctx context.Context) (components.ObjectList[LookupLogDisplay], error) {
		raw := ctx.Value(ctxKeyLookupLogEntries)
		if raw == nil {
			return components.ObjectList[LookupLogDisplay]{Items: nil, Total: 0, Number: 1, NumPages: 1}, nil
		}
		displays, ok := raw.([]LookupLogDisplay)
		if !ok {
			return components.ObjectList[LookupLogDisplay]{Items: nil, Total: 0, Number: 1, NumPages: 1}, nil
		}
		n := uint64(len(displays))
		return components.ObjectList[LookupLogDisplay]{
			Items:    displays,
			Number:   1,
			NumPages: 1,
			Total:    n,
		}, nil
	}
}

func lookupLogRowMarkdownGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		row, ok := ctx.Value("$row").(map[string]any)
		if !ok {
			return "", nil
		}
		kindKey, _ := row["Kind"].(string)
		label := kindKey
		if p, ok := registry.PairFromPairs(kindKey, LookupLogEntryKindChoices); ok {
			label = p.Value
		}
		var ts string
		if t, ok := row["CreatedAt"].(time.Time); ok && !t.IsZero() {
			tz, _ := ctx.Value("$tz").(*time.Location)
			if tz == nil {
				tz = time.UTC
			}
			ts = t.In(tz).Format(time.RFC3339)
		}
		var id uint
		if v, ok := row["ID"].(uint); ok {
			id = v
		}
		header := fmt.Sprintf("**Kind:** %s · **Created:** %s · **ID:** %d\n\n", label, ts, id)

		var body string
		switch kindKey {
		case "thought":
			if th, ok := row["Thought"].(*LookupThought); ok && th != nil {
				body = strings.TrimSpace(th.Text)
			}
			if body == "" {
				body = "_(no thought body)_"
			}
		case "text":
			if tx, ok := row["LogText"].(*LookupText); ok && tx != nil {
				body = strings.TrimSpace(tx.Text)
			}
			if body == "" {
				body = "_(no text body)_"
			}
		case "tool_call":
			tc, _ := row["ToolCall"].(*LookupToolCall)
			if tc == nil {
				body = "_(no tool call payload)_"
			} else {
				var b strings.Builder
				fmt.Fprintf(&b, "**Tool:** `%s`\n\n", tc.Name)
				b.WriteString("**Arguments**\n\n```json\n")
				b.WriteString(prettyDatatypesJSON(tc.Arguments))
				b.WriteString("\n```\n\n**Result**\n\n```json\n")
				b.WriteString(prettyDatatypesJSON(tc.Result))
				b.WriteString("\n```")
				body = b.String()
			}
		case "tool_error":
			te, _ := row["ToolError"].(*LookupToolError)
			if te == nil {
				body = "_(no tool error payload)_"
			} else {
				var b strings.Builder
				fmt.Fprintf(&b, "**Tool:** `%s`\n\n**Message:** %s\n\n", te.ToolName, te.Message)
				if len(te.Detail) > 0 {
					b.WriteString("**Detail**\n\n```json\n")
					b.WriteString(prettyDatatypesJSON(te.Detail))
					b.WriteString("\n```")
				}
				body = b.String()
			}
		default:
			body = fmt.Sprintf("_(unknown kind %q — raw kind key: %q)_", label, kindKey)
		}

		return header + body, nil
	}
}

func prettyDatatypesJSON(j datatypes.JSON) string {
	if len(j) == 0 {
		return "(none)"
	}
	var out bytes.Buffer
	if err := json.Indent(&out, j, "", "  "); err != nil {
		slog.Warn("lacerate: lookup log pretty json indent", "error", err)
		return string(j)
	}
	return out.String()
}

func lookupDetailLogSection() components.PageInterface {
	return &components.Timeline[LookupLogDisplay]{
		Page:    components.Page{Key: "lacerate.LookupDetailLogs"},
		UID:     "lacerate-lookup-log-timeline",
		Title:   "Activity log",
		Classes: "w-full mt-8",
		Data:    lookupLogEntriesObjectListGetter(),
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page: components.Page{Key: "lacerate.LookupDetailLogRow"},
				Children: []components.PageInterface{
					&components.FieldMarkdown{
						Getter:  lookupLogRowMarkdownGetter(),
						Classes: "prose prose-sm max-w-none",
					},
				},
			},
		},
	}
}
