package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/getters"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
)

// ProgramDisplayLabel returns "Name (University label)" using [UniversityChoices]; empty university key → name only.
func ProgramDisplayLabel(nameGetter, universityKeyGetter getters.Getter[string]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		name, err := nameGetter(ctx)
		if err != nil {
			return "", err
		}
		ukey, errU := universityKeyGetter(ctx)
		if errU != nil || ukey == "" {
			return name, nil
		}
		if p, ok := registry.PairFromPairs(ukey, UniversityChoices); ok {
			return fmt.Sprintf("%s (%s)", name, p.Value), nil
		}
		return fmt.Sprintf("%s (%s)", name, ukey), nil
	}
}

func courseListDisplayGetter(g getters.Getter[[]courses.Course]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		list, err := g(ctx)
		if err != nil {
			return "", err
		}
		if len(list) == 0 {
			return "", nil
		}
		codes := make([]string, 0, len(list))
		for _, c := range list {
			codes = append(codes, c.Code)
		}
		return strings.Join(codes, ", "), nil
	}
}
