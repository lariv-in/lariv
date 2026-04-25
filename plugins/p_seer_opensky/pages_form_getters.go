package p_seer_opensky

import (
	"context"
	"strconv"

	"github.com/lariv-in/lago/getters"
)

func int64PtrForm(g getters.Getter[*int64]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.FormatInt(*p, 10), nil
	}
}

func intPtrForm(g getters.Getter[*int]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.Itoa(*p), nil
	}
}

func floatPtrForm(g getters.Getter[*float64]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.FormatFloat(*p, 'f', 6, 64), nil
	}
}

func boolPtrForm(g getters.Getter[*bool]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		if *p {
			return "true", nil
		}
		return "false", nil
	}
}
