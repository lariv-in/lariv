package p_totschool_proposals

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
)

func getterGenerated() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content != "" {
			return true, nil
		}
		return false, nil
	}
}

func getterGenerationPending() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id != nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}

func getterIdleGeneration() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		id, err := getters.Key[*int]("$in.GenerationID")(ctx)
		if err != nil {
			slog.Error("Error while getting id for checking if appointment is idle", "error", err)
			return false, err
		}
		content, err := getters.Key[string]("$in.GeneratedContent")(ctx)
		if err != nil {
			slog.Error("Error while getting content for checking if appointment is idle", "error", err)
			return false, err
		}
		if id == nil && content == "" {
			return true, nil
		}
		return false, nil
	}
}
