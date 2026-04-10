package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"gorm.io/datatypes"
)

var _ components.InputInterface = InputSubredditList{}

// InputSubredditList embeds [components.InputStringList] and normalizes subreddit names in Parse (trim, strip optional leading r/).
// Parse returns [datatypes.JSON] so form values match the model field without extra conversion in view layers.
type InputSubredditList struct {
	components.InputStringList
}

func (e InputSubredditList) Parse(v any, ctx context.Context) (any, error) {
	rawAny, err := e.InputStringList.Parse(v, ctx)
	if err != nil {
		return nil, err
	}
	s, ok := rawAny.(string)
	if !ok {
		return rawAny, nil
	}
	if s == "[]" {
		return datatypes.JSON([]byte("[]")), nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, fmt.Errorf("subreddits must be a JSON array of strings: %w", err)
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		x := strings.TrimSpace(item)
		if len(x) >= 2 && strings.EqualFold(x[:2], "r/") {
			x = strings.TrimSpace(x[2:])
		}
		if x != "" {
			out = append(out, x)
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}
