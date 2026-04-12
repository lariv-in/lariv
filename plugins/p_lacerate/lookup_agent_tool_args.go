package p_lacerate

import (
	"encoding/json"
	"fmt"
)

// unmarshalToolArgs re-encodes GenAI tool args (map[string]any) as JSON and decodes into dst.
// Round-tripping normalizes types and applies standard json struct rules.
func unmarshalToolArgs(args map[string]any, dst any) error {
	if args == nil {
		args = map[string]any{}
	}
	b, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal tool args: %w", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("unmarshal tool args: %w", err)
	}
	return nil
}

type createReportArgs struct {
	Name            string                    `json:"name"`
	TargetKind      string                    `json:"target_kind"`
	Description     *string                   `json:"description"`
	BriefingContent *string                   `json:"briefing_content"`
	TimelineEntries []reportTimelineEntryArgs `json:"timeline_entries"`
}

type editReportArgs struct {
	ID              uint                       `json:"id"`
	Name            *string                    `json:"name"`
	TargetKind      *string                    `json:"target_kind"`
	Description     *string                    `json:"description"`
	BriefingContent *string                    `json:"briefing_content"`
	TimelineEntries *[]reportTimelineEntryArgs `json:"timeline_entries"`
}

type appendTimelineEntriesArgs struct {
	ID              uint                      `json:"id"`
	TimelineEntries []reportTimelineEntryArgs `json:"timeline_entries"`
}

type reportTimelineEntryArgs struct {
	Datetime string `json:"datetime"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type createTargetOfInterestArgs struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type editTargetOfInterestArgs struct {
	ID          uint    `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type embeddingSearchArgs struct {
	Query string `json:"query"`
	Limit *int   `json:"limit"`
}

type attachEventArgs struct {
	IntelID  uint   `json:"intel_id"`
	Datetime string `json:"datetime"`
	Address  string `json:"address"`
}

type removeEventArgs struct {
	ID uint `json:"id"`
}

func parseLookupSearchLimit(limit *int, defaultLimit, max int) (int, error) {
	if limit == nil {
		return defaultLimit, nil
	}
	n := *limit
	if n < 1 || n > max {
		return 0, fmt.Errorf("limit must be between 1 and %d", max)
	}
	return n, nil
}
