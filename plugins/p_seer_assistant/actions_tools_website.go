package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_seer_websites"
	"gorm.io/gorm"
)

func runWebsiteListSourcesTool(ctx context.Context, db *gorm.DB, _ *assistantToolEnvelope) (string, string) {
	s, err := p_seer_websites.ListWebsiteSourcesJSON(ctx, db)
	if err != nil {
		return "", err.Error()
	}
	return s, ""
}

func runWebsiteListWorkersTool(ctx context.Context, db *gorm.DB, _ *assistantToolEnvelope) (string, string) {
	s, err := p_seer_websites.ListWebsiteRunnersJSON(ctx, db)
	if err != nil {
		return "", err.Error()
	}
	return s, ""
}

func runWebsiteAddSourceTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	seed := strings.TrimSpace(env.SeedURL)
	if seed == "" {
		return "", "website_add_source: seed_url required"
	}
	src, err := p_seer_websites.CreateWebsiteSourceFromParams(ctx, tx, p_seer_websites.WebsiteSourceCreateParams{
		SeedURL:         seed,
		Depth:           env.WebsiteDepth,
		WebsiteRunnerID: env.WebsiteRunnerFK,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_source_id": src.ID,
		"seed_url":          strings.TrimSpace(src.URL.String()),
		"website_depth":     src.Depth,
	}
	if src.WebsiteRunnerID != nil {
		out["website_runner_id"] = *src.WebsiteRunnerID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteEditSourceTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	if env.WebsiteSourceID == 0 {
		return "", "website_edit_source: website_source_id required"
	}
	seed := strings.TrimSpace(env.SeedURL)
	if seed == "" {
		return "", "website_edit_source: seed_url required"
	}
	src, err := p_seer_websites.UpdateWebsiteSourceFromParams(ctx, tx, p_seer_websites.WebsiteSourceUpdateParams{
		SourceID:        env.WebsiteSourceID,
		SeedURL:         seed,
		Depth:           env.WebsiteDepth,
		WebsiteRunnerID: env.WebsiteRunnerFK,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_source_id": src.ID,
		"seed_url":          strings.TrimSpace(src.URL.String()),
		"website_depth":     src.Depth,
	}
	if src.WebsiteRunnerID != nil {
		out["website_runner_id"] = *src.WebsiteRunnerID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteAddWorkerTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	name := strings.TrimSpace(env.WorkerName)
	durStr := strings.TrimSpace(env.WorkerDuration)
	if name == "" {
		return "", "website_add_worker: worker_name required"
	}
	if durStr == "" {
		return "", "website_add_worker: worker_duration required (Go duration, e.g. 1h, 30m)"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("website_add_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_websites.CreateWebsiteRunnerFromParams(ctx, tx, p_seer_websites.WebsiteRunnerCreateParams{
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_runner_id": runner.ID,
		"name":              runner.Name,
		"duration":          runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}

func runWebsiteEditWorkerTool(ctx context.Context, tx *gorm.DB, env *assistantToolEnvelope) (string, string) {
	if env.WebsiteRunnerFK == nil || *env.WebsiteRunnerFK == 0 {
		return "", "website_edit_worker: website_runner_id required"
	}
	name := strings.TrimSpace(env.WorkerName)
	durStr := strings.TrimSpace(env.WorkerDuration)
	if name == "" {
		return "", "website_edit_worker: worker_name required"
	}
	if durStr == "" {
		return "", "website_edit_worker: worker_duration required"
	}
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return "", fmt.Sprintf("website_edit_worker: invalid worker_duration: %v", err)
	}
	runner, err := p_seer_websites.UpdateWebsiteRunnerFromParams(ctx, tx, p_seer_websites.WebsiteRunnerUpdateParams{
		ID:       *env.WebsiteRunnerFK,
		Name:     name,
		Duration: d,
	})
	if err != nil {
		return "", err.Error()
	}
	out := map[string]any{
		"website_runner_id": runner.ID,
		"name":              runner.Name,
		"duration":          runner.Duration.String(),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err.Error()
	}
	return string(b), ""
}
