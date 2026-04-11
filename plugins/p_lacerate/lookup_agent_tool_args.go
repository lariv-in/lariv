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
	Name        string  `json:"name"`
	TargetType  string  `json:"target_type"`
	Description *string `json:"description"`
	Content     string  `json:"content"`
}

type editReportArgs struct {
	ID          uint    `json:"id"`
	Name        *string `json:"name"`
	TargetType  *string `json:"target_type"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
}

type embeddingSearchArgs struct {
	Query string `json:"query"`
	Limit *int   `json:"limit"`
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
