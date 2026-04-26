package p_seer_gdelt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
)

func gdeltPairGetterWithDefault(field string, choices []registry.Pair[string, string], fallback string) getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$get."+field)(ctx)
		if err != nil {
			return registry.Pair[string, string]{}, nil
		}
		s = strings.TrimSpace(s)
		if s == "" {
			s = fallback
		}
		if s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromPairs(s, choices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func gdeltDateGetter(field string) getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		raw, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return time.Time{}, nil
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return time.Time{}, nil
		}
		t, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			return time.Time{}, nil
		}
		return t, nil
	}
}

func gdeltUintGetter(field string, fallback uint) getters.Getter[uint] {
	return func(ctx context.Context) (uint, error) {
		raw, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return fallback, nil
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return fallback, nil
		}
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fallback, nil
		}
		return uint(n), nil
	}
}

func gdeltActorsGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		actor1, _ := getters.Key[string]("$row.Actor1Name")(ctx)
		actor2, _ := getters.Key[string]("$row.Actor2Name")(ctx)
		actor1 = strings.TrimSpace(actor1)
		actor2 = strings.TrimSpace(actor2)
		switch {
		case actor1 != "" && actor2 != "":
			return actor1 + " / " + actor2, nil
		case actor1 != "":
			return actor1, nil
		case actor2 != "":
			return actor2, nil
		default:
			return "Unknown actors", nil
		}
	}
}

func gdeltEventGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		eventCode, _ := getters.Key[string]("$row.EventCode")(ctx)
		action, _ := getters.Key[string]("$row.ActionGeoFullName")(ctx)
		eventCode = strings.TrimSpace(eventCode)
		action = strings.TrimSpace(action)
		switch {
		case eventCode != "" && action != "":
			return eventCode + " in " + action, nil
		case eventCode != "":
			return eventCode, nil
		case action != "":
			return action, nil
		default:
			return "Event", nil
		}
	}
}

func gdeltSourceLabelGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		url, _ := getters.Key[string]("$row.SourceURL")(ctx)
		url = strings.TrimSpace(url)
		if url == "" {
			return "", nil
		}
		normalized := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://"), "www.")
		if idx := strings.Index(normalized, "/"); idx >= 0 {
			normalized = normalized[:idx]
		}
		if normalized != "" {
			return normalized, nil
		}
		return url, nil
	}
}

func gdeltRowDateGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, err := getters.Key[int]("$row.SQLDate")(ctx)
		if err != nil || n == 0 {
			return "", nil
		}
		s := strconv.Itoa(n)
		if len(s) != 8 {
			return s, nil
		}
		return s[:4] + "-" + s[4:6] + "-" + s[6:], nil
	}
}

func gdeltRowMentionsGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, err := getters.Key[int]("$row.NumMentions")(ctx)
		if err != nil {
			return "", nil
		}
		return strconv.Itoa(n), nil
	}
}

func gdeltFmtAny(path string) getters.Getter[string] {
	return getters.Map(getters.Key[any](path), func(_ context.Context, v any) (string, error) {
		return fmt.Sprint(v), nil
	})
}

func gdeltListSQLDateGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, err := getters.Key[int]("$row.SQLDate")(ctx)
		if err != nil || n == 0 {
			return "", err
		}
		s := strconv.Itoa(n)
		if len(s) != 8 {
			return s, nil
		}
		return s[:4] + "-" + s[4:6] + "-" + s[6:], nil
	}
}

func gdeltListActorsGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		a1, _ := getters.Key[string]("$row.Actor1Name")(ctx)
		a2, _ := getters.Key[string]("$row.Actor2Name")(ctx)
		a1 = strings.TrimSpace(a1)
		a2 = strings.TrimSpace(a2)
		switch {
		case a1 != "" && a2 != "":
			return a1 + " / " + a2, nil
		case a1 != "":
			return a1, nil
		case a2 != "":
			return a2, nil
		default:
			return "", nil
		}
	}
}
