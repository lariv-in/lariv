package p_seer_deepsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/plugins/p_seer_websites"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const (
	maxTotalScrapeURLs        = 30
	maxGeneratedSearchQueries = 8
	// deepSearchScrapeConcurrency limits parallel headless scrapes (shared browser / IO).
	deepSearchScrapeConcurrency = 4
	// deepSearchIntelConcurrency limits parallel local-model work during website→Intel ingest.
	deepSearchIntelConcurrency = 3
)

func persistDeepSearch(ctx context.Context, db *gorm.DB, id uint, fields map[string]any) error {
	return db.WithContext(ctx).Model(&DeepSearch{}).Where("id = ?", id).Updates(fields).Error
}

// deepSearchIntelIngestOne runs Intel ingest for one [p_seer_websites.Website] and appends [DeepSearchLog] lines.
func deepSearchIntelIngestOne(ctx context.Context, db *gorm.DB, deepSearchID uint, w p_seer_websites.Website, kind string) {
	had, errEx := p_seer_intel.IntelExistsForSource(ctx, db, kind, w.ID)
	if errEx != nil {
		appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindError, fmt.Sprintf("intel exists check website_id=%d: %v", w.ID, errEx))
		slog.Warn("p_seer_deepsearch: intel exists", "website_id", w.ID, "error", errEx)
		return
	}
	p_seer_websites.RunWebsiteSingleIntelIngest(ctx, db, w)
	has, errAfter := p_seer_intel.IntelExistsForSource(ctx, db, kind, w.ID)
	if errAfter != nil {
		appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindError, fmt.Sprintf("intel exists re-check website_id=%d: %v", w.ID, errAfter))
		return
	}
	if had {
		appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindIntelUnchanged, fmt.Sprintf("website_id=%d url=%s (intel already existed)", w.ID, w.URL.String()))
		return
	}
	if has {
		var in p_seer_intel.Intel
		if err := db.WithContext(ctx).Where("kind = ? AND kind_id = ? AND deleted_at IS NULL", kind, w.ID).
			Order("id DESC").First(&in).Error; err != nil {
			appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindError, fmt.Sprintf("intel row lookup website_id=%d: %v", w.ID, err))
			return
		}
		appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindIntelCreated, fmt.Sprintf("intel_id=%d website_id=%d url=%s", in.ID, w.ID, w.URL.String()))
		return
	}
	appendDeepSearchLog(ctx, db, deepSearchID, DeepSearchLogKindIntelCreateFailed, fmt.Sprintf("website_id=%d url=%s (no intel row after ingest)", w.ID, w.URL.String()))
}

func runDeepSearchPipeline(ctx context.Context, db *gorm.DB, id uint) {
	// #region agent log
	p_seer_intel.AgentDebugSessionLog("H3", "actions_deepsearch.go:runDeepSearchPipeline", "pipeline_start", map[string]any{
		"deep_search_id": id,
	})
	// #endregion
	var row DeepSearch
	if err := db.WithContext(ctx).First(&row, id).Error; err != nil {
		slog.Error("p_seer_deepsearch: load row", "id", id, "error", err)
		return
	}
	userQuery := strings.TrimSpace(row.Query)
	if userQuery == "" {
		appendDeepSearchLog(ctx, db, id, DeepSearchLogKindError, "empty user query")
		_ = persistDeepSearch(ctx, db, id, map[string]any{
			"status":    DeepSearchStatusFailed,
			"run_error": "empty query",
		})
		return
	}

	fail := func(msg string, err error) bool {
		if err != nil && errors.Is(err, context.Canceled) {
			deepSearchFinishCancelled(ctx, db, id)
			return true
		}
		s := msg
		if err != nil {
			s = msg + ": " + err.Error()
		}
		slog.Error("p_seer_deepsearch: pipeline", "deep_search_id", id, "error", s)
		appendDeepSearchLog(ctx, db, id, DeepSearchLogKindError, s)
		_ = persistDeepSearch(ctx, db, id, map[string]any{
			"status":    DeepSearchStatusFailed,
			"run_error": s,
		})
		return true
	}

	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusRunning}); err != nil {
		slog.Error("p_seer_deepsearch: persist running", "id", id, "error", err)
		return
	}
	appendDeepSearchLog(ctx, db, id, DeepSearchLogKindInfo, "pipeline started")

	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusExpandingQueries}); err != nil {
		if fail("persist status", err) {
			return
		}
	}
	if deepSearchAbortIfCtxDone(ctx, db, id) {
		return
	}
	queries, err := expandDeepSearchQueries(ctx, userQuery)
	if err != nil {
		if fail("expand queries", err) {
			return
		}
	}
	if b, jerr := json.MarshalIndent(queries, "", "  "); jerr == nil {
		appendDeepSearchLog(ctx, db, id, DeepSearchLogKindQueriesGenerated, string(b))
	} else {
		appendDeepSearchLog(ctx, db, id, DeepSearchLogKindQueriesGenerated, strings.Join(queries, "\n"))
	}

	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusSearching}); err != nil {
		if fail("persist status", err) {
			return
		}
	}
	seen := make(map[string]struct{})
	var urlList []*url.URL
	for _, q := range queries {
		if deepSearchAbortIfCtxDone(ctx, db, id) {
			return
		}
		if len(urlList) >= maxTotalScrapeURLs {
			break
		}
		links, gerr := googleCustomSearchURLs(ctx, q)
		if gerr != nil {
			appendDeepSearchLog(ctx, db, id, DeepSearchLogKindError, fmt.Sprintf("Google CSE query %q failed: %v", q, gerr))
			slog.Warn("p_seer_deepsearch: CSE query failed", "query", q, "error", gerr)
			continue
		}
		appendDeepSearchLog(ctx, db, id, DeepSearchLogKindSearchPerformed, fmt.Sprintf("query=%q unique_links_this_page=%d", q, len(links)))
		for _, raw := range links {
			if len(urlList) >= maxTotalScrapeURLs {
				break
			}
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			u, perr := url.Parse(raw)
			if perr != nil || u.Host == "" {
				continue
			}
			switch strings.ToLower(u.Scheme) {
			case "http", "https":
			default:
				continue
			}
			key := u.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			urlList = append(urlList, u)
		}
	}

	if deepSearchAbortIfCtxDone(ctx, db, id) {
		return
	}
	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusScraping}); err != nil {
		if fail("persist status", err) {
			return
		}
	}
	scrapeSem := make(chan struct{}, deepSearchScrapeConcurrency)
	scrapeG, scrapeCtx := errgroup.WithContext(ctx)
	for _, u := range urlList {
		u := u
		scrapeG.Go(func() error {
			scrapeSem <- struct{}{}
			defer func() { <-scrapeSem }()
			if err := p_seer_websites.WebsiteScrapeIfAbsent(scrapeCtx, db, u); err != nil {
				appendDeepSearchLog(scrapeCtx, db, id, DeepSearchLogKindError, fmt.Sprintf("scrape %s: %v", u.String(), err))
				slog.Warn("p_seer_deepsearch: scrape", "url", u.String(), "error", err)
				return nil
			}
			appendDeepSearchLog(scrapeCtx, db, id, DeepSearchLogKindWebsiteFetched, u.String())
			return nil
		})
	}
	_ = scrapeG.Wait()

	if deepSearchAbortIfCtxDone(ctx, db, id) {
		return
	}
	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusIngestingIntel}); err != nil {
		if fail("persist status", err) {
			return
		}
	}
	sitesByID := make(map[uint]p_seer_websites.Website)
	for _, u := range urlList {
		w, werr := p_seer_websites.WebsiteByFetchableURL(ctx, db, u)
		if werr != nil || w.ID == 0 {
			continue
		}
		sitesByID[w.ID] = w
	}
	kind := (p_seer_websites.Website{}).Kind()
	siteIDs := make([]uint, 0, len(sitesByID))
	for wid := range sitesByID {
		siteIDs = append(siteIDs, wid)
	}
	sort.Slice(siteIDs, func(i, j int) bool { return siteIDs[i] < siteIDs[j] })

	intelSem := make(chan struct{}, deepSearchIntelConcurrency)
	intelG, intelCtx := errgroup.WithContext(ctx)
	for _, wid := range siteIDs {
		w := sitesByID[wid]
		intelG.Go(func() error {
			intelSem <- struct{}{}
			defer func() { <-intelSem }()
			deepSearchIntelIngestOne(intelCtx, db, id, w, kind)
			return nil
		})
	}
	_ = intelG.Wait()

	if deepSearchAbortIfCtxDone(ctx, db, id) {
		return
	}
	if err := persistDeepSearch(ctx, db, id, map[string]any{"status": DeepSearchStatusReporting}); err != nil {
		if fail("persist status", err) {
			return
		}
	}
	appendDeepSearchLog(ctx, db, id, DeepSearchLogKindInfo, "generating report (LLM with Intel tools)")
	report, err := runDeepSearchReport(ctx, db, id, userQuery)
	if err != nil {
		if fail("report", err) {
			return
		}
	}

	if err := persistDeepSearch(ctx, db, id, map[string]any{
		"status":    DeepSearchStatusDone,
		"report":    report,
		"run_error": "",
	}); err != nil {
		slog.Error("p_seer_deepsearch: persist done", "id", id, "error", err)
		return
	}
	appendDeepSearchLog(ctx, db, id, DeepSearchLogKindInfo, fmt.Sprintf("pipeline finished; report length=%d runes", len([]rune(report))))
}
