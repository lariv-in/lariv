package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
)

var _ components.InputInterface = InputTwitterHandleList{}

// InputTwitterHandleList embeds [components.InputStringList] and normalizes Twitter handles in Parse (trim, strip one leading @).
type InputTwitterHandleList struct {
	components.InputStringList
}

func (e InputTwitterHandleList) Parse(v any, ctx context.Context) (any, error) {
	rawAny, err := e.InputStringList.Parse(v, ctx)
	if err != nil {
		return nil, err
	}
	s, ok := rawAny.(string)
	if !ok {
		return rawAny, nil
	}
	if s == "[]" {
		return s, nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, fmt.Errorf("handles must be a JSON array of strings: %w", err)
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		x := strings.TrimSpace(item)
		if strings.HasPrefix(x, "@") {
			x = strings.TrimSpace(x[1:])
		}
		if x != "" {
			out = append(out, x)
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
