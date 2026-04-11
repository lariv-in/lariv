package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/lariv-in/lago/registry"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const lookupAgentMaxIters = 24

// firstViableLookupCandidate returns the first non-nil candidate with non-nil Content.
func firstViableLookupCandidate(cands []*genai.Candidate) *genai.Candidate {
	for _, c := range cands {
		if c != nil && c.Content != nil {
			return c
		}
	}
	return nil
}

// functionCallsFromContent collects function calls from one candidate's content (same logic as
// [genai.GenerateContentResponse.FunctionCalls], but not tied to Candidates[0]).
func functionCallsFromContent(content *genai.Content) []*genai.FunctionCall {
	if content == nil || len(content.Parts) == 0 {
		return nil
	}
	var out []*genai.FunctionCall
	for _, part := range content.Parts {
		if part != nil && part.FunctionCall != nil {
			out = append(out, part.FunctionCall)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

const osintSpecialistSystem = `You are an OSINT Specialist: methodical, careful, and ethical. You only use the provided tools to gather and record intelligence from this application's corpora. Prefer verifiable facts from tool results over speculation.

Actively work with targets of interest (TOI) as part of every lookup: use get_relevant_targets_of_interest when the briefing or intel involves specific people, organizations, places, or other entities, so you know what the user already tracks; use create_target_of_interest and edit_target_of_interest when warranted below. Treat TOIs as a first-class store alongside reports and ingested intel, not an afterthought.

Targets of interest are very short, accurate descriptions of entities the user cares about. Create or edit a TOI only when you find information that materially changes the context of that entity—not for minor or redundant intel. Before creating a new TOI, search with get_relevant_targets_of_interest to avoid duplicates and to pick the right row to edit.

For reports: strongly prefer editing an existing report with edit_report over creating a new report with create_report when the same subject is already covered. Before creating a new report, search for an existing one for the same subject; if one already exists, update it instead of creating a duplicate. For timeline reports, edit_report.timeline_entries replaces the whole timeline; use append_timeline_entries when you only need to add new events.

When done, give a concise summary of what you did and what you learned.`

func runLookupAgent(ctx context.Context, db *gorm.DB, lu *Lookup) error {
	if lu == nil {
		err := fmt.Errorf("nil lookup")
		slog.Error("lacerate: lookup agent", "error", err)
		return err
	}
	key := strings.TrimSpace(Config.GeminiEmbedding.APIKey)
	if key == "" {
		slog.Info("lacerate: lookup agent skipped (no gemini api key)", "lookup_id", lu.ID)
		return nil
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		err = fmt.Errorf("genai client: %w", err)
		slog.Error("lacerate: lookup agent genai client", "error", err, "lookup_id", lu.ID)
		return err
	}
	model := strings.TrimSpace(Config.GeminiAgent.Model)
	if model == "" {
		model = "gemini-2.5-flash"
	}

	run := &lookupRun{db: db, lookupID: lu.ID}
	task := fmt.Sprintf("Execute this OSINT lookup task using tools as needed.\n\n## Lookup briefing\n%s", lu.Content)

	contents := []*genai.Content{
		genai.NewContentFromText(task, genai.RoleUser),
	}

	cfg := &genai.GenerateContentConfig{
		Temperature:       genai.Ptr[float32](0.2),
		SystemInstruction: genai.NewContentFromText(osintSpecialistSystem, genai.RoleUser),
		Tools:             []*genai.Tool{{FunctionDeclarations: lookupAgentToolDeclarations()}},
	}

	for range lookupAgentMaxIters {
		if err := ctx.Err(); err != nil {
			slog.Error("lacerate: lookup agent context", "error", err, "lookup_id", lu.ID)
			return err
		}
		resp, err := client.Models.GenerateContent(ctx, model, contents, cfg)
		if err != nil {
			err = fmt.Errorf("generate: %w", err)
			slog.Error("lacerate: lookup agent generate", "error", err, "lookup_id", lu.ID)
			return err
		}
		cand := firstViableLookupCandidate(resp.Candidates)
		if cand == nil {
			slog.Error("lacerate: lookup agent: empty candidate", "lookup_id", lu.ID)
			break
		}
		run.logModelParts(cand.Content.Parts)

		calls := functionCallsFromContent(cand.Content)
		if len(calls) == 0 {
			break
		}

		contents = append(contents, cand.Content)

		var frParts []*genai.Part
		for _, fc := range calls {
			if fc == nil {
				continue
			}
			name := strings.TrimSpace(fc.Name)
			args := fc.Args

			toolRow, logErr := run.startToolCall(name, args)
			if logErr != nil {
				slog.Error("lacerate: lookup agent log tool call", "error", logErr, "lookup_id", lu.ID, "tool", name)
			}

			out, toolErr := run.dispatchTool(ctx, name, args)
			fr := &genai.FunctionResponse{Name: name, ID: fc.ID}
			if toolErr != nil {
				if err2 := run.recordToolError(name, toolErr.Error(), map[string]any{"args": args}); err2 != nil {
					slog.Error("lacerate: lookup agent log tool error", "error", err2, "lookup_id", lu.ID)
				}
				fr.Response = map[string]any{"error": toolErr.Error()}
			} else {
				if toolRow != nil {
					if err2 := run.finishToolCall(toolRow, out); err2 != nil {
						slog.Error("lacerate: lookup agent save tool result", "error", err2, "lookup_id", lu.ID)
					}
				}
				fr.Response = map[string]any{"output": out}
			}
			frParts = append(frParts, &genai.Part{FunctionResponse: fr})
		}

		if len(frParts) == 0 {
			break
		}
		contents = append(contents, genai.NewContentFromParts(frParts, genai.RoleUser))
	}

	return nil
}

// lookupAgentTool is one Gemini-callable tool for the OSINT lookup agent.
type lookupAgentTool interface {
	Name() string
	Declaration() *genai.FunctionDeclaration
	Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error)
}

// lookupAgentTools is the single registry for lookup-agent tools (name → tool, stable order for declarations).
var lookupAgentTools = registry.NewRegistry[lookupAgentTool]()

func init() {
	for _, t := range []lookupAgentTool{
		createReportTool{},
		editReportTool{},
		appendTimelineEntriesTool{},
		getRelevantReportsTool{},
		getRelevantIntelTool{},
		createTargetOfInterestTool{},
		editTargetOfInterestTool{},
		getRelevantTargetsOfInterestTool{},
	} {
		if err := lookupAgentTools.Register(t.Name(), t); err != nil {
			panic(err)
		}
	}
}

func lookupAgentToolDeclarations() []*genai.FunctionDeclaration {
	pairs := *lookupAgentTools.AllStable(registry.AlphabeticalByKey[lookupAgentTool]{})
	out := make([]*genai.FunctionDeclaration, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.Value.Declaration())
	}
	return out
}

type lookupRun struct {
	db       *gorm.DB
	lookupID uint
}

func (r *lookupRun) logModelParts(parts []*genai.Part) {
	for _, p := range parts {
		if p == nil {
			continue
		}
		if p.FunctionCall != nil {
			continue
		}
		t := strings.TrimSpace(p.Text)
		if t == "" {
			continue
		}
		if p.Thought {
			th := &LookupThought{Text: t}
			if err := CreateLookupLogEntryData(r.db, r.lookupID, th); err != nil {
				slog.Error("lacerate: lookup log thought", "error", err, "lookup_id", r.lookupID)
			}
			continue
		}
		lt := &LookupText{Text: t}
		if err := CreateLookupLogEntryData(r.db, r.lookupID, lt); err != nil {
			slog.Error("lacerate: lookup log text", "error", err, "lookup_id", r.lookupID)
		}
	}
}

func (r *lookupRun) startToolCall(name string, args map[string]any) (*LookupToolCall, error) {
	argJSON, err := json.Marshal(args)
	if err != nil {
		slog.Error("lacerate: lookup tool call marshal args", "error", err, "lookup_id", r.lookupID, "tool", name)
		return nil, err
	}
	tool := &LookupToolCall{
		Name:      name,
		Arguments: datatypes.JSON(argJSON),
	}
	if err := CreateLookupLogEntryData(r.db, r.lookupID, tool); err != nil {
		slog.Error("lacerate: lookup tool call create", "error", err, "lookup_id", r.lookupID, "tool", name)
		return nil, err
	}
	return tool, nil
}

func (r *lookupRun) finishToolCall(row *LookupToolCall, result any) error {
	if row == nil || row.ID == 0 {
		return nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		slog.Error("lacerate: lookup tool finish marshal", "error", err, "lookup_id", r.lookupID)
		return err
	}
	if err := r.db.Model(row).Update("result", datatypes.JSON(b)).Error; err != nil {
		slog.Error("lacerate: lookup tool finish update", "error", err, "lookup_id", r.lookupID)
		return err
	}
	return nil
}

func (r *lookupRun) recordToolError(toolName, message string, detail any) error {
	var det datatypes.JSON
	if detail != nil {
		b, err := json.Marshal(detail)
		if err != nil {
			slog.Error("lacerate: lookup tool error marshal detail", "error", err, "lookup_id", r.lookupID)
			return err
		}
		det = b
	}
	errRow := &LookupToolError{
		ToolName: toolName,
		Message:  message,
		Detail:   det,
	}
	if err := CreateLookupLogEntryData(r.db, r.lookupID, errRow); err != nil {
		slog.Error("lacerate: lookup tool error create", "error", err, "lookup_id", r.lookupID)
		return err
	}
	return nil
}

func (r *lookupRun) dispatchTool(ctx context.Context, name string, args map[string]any) (any, error) {
	t, ok := lookupAgentTools.Get(name)
	if !ok {
		err := fmt.Errorf("unknown tool %q", name)
		slog.Warn("lacerate: lookup agent dispatch", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return t.Run(ctx, r, args)
}

type createReportTool struct{}

func (createReportTool) Name() string { return "create_report" }

func (createReportTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "create_report",
		Description: "Create a curated report in the database. Before creating, check for an existing relevant report and use edit_report instead if one already exists.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":             map[string]any{"type": "string", "description": "Short title for the report."},
				"target_kind":      map[string]any{"type": "string", "description": "One of: briefing, timeline."},
				"description":      map[string]any{"type": "string", "description": "Summary or context."},
				"briefing_content": map[string]any{"type": "string", "description": "Required when target_kind is briefing. Large markdown body for the briefing."},
				"timeline_entries": map[string]any{
					"type":        "array",
					"description": "Required when target_kind is timeline. Timeline entries in chronological order.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"datetime": map[string]any{"type": "string", "description": "Entry datetime in RFC3339."},
							"title":    map[string]any{"type": "string", "description": "Short entry title."},
							"content":  map[string]any{"type": "string", "description": "Markdown body for the entry."},
						},
						"required": []string{"datetime", "title", "content"},
					},
				},
			},
			"required": []string{"name", "target_kind"},
		},
	}
}

func (createReportTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p createReportArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	name := strings.TrimSpace(p.Name)
	kind := strings.TrimSpace(p.TargetKind)
	desc := ""
	if p.Description != nil {
		desc = strings.TrimSpace(*p.Description)
	}
	if name == "" {
		err := fmt.Errorf("name is required")
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if !validReportKind(kind) {
		err := fmt.Errorf("target_kind must be one of: briefing, timeline")
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	formData, err := reportFormDataFromToolArgs(ctx, kind, p.BriefingContent, p.TimelineEntries)
	if err != nil {
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	formData.Name = name
	formData.Description = desc
	var report Report
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		report = Report{Name: name, Description: desc, Kind: kind}
		if err := tx.Create(&report).Error; err != nil {
			return err
		}
		if err := createReportKindRow(tx, report.ID, formData); err != nil {
			return err
		}
		return refreshReportEmbedding(ctx, tx, report.ID)
	}); err != nil {
		slog.Error("lacerate: lookup tool create_report db", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": report.ID, "name": report.Name, "kind": report.Kind}, nil
}

type editReportTool struct{}

func (editReportTool) Name() string { return "edit_report" }

func (editReportTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "edit_report",
		Description: "Update an existing report by ID. Use when an existing report covers the same subject and should be revised instead of creating a duplicate. Only include scalar fields you want to change; if you provide timeline_entries, they replace the entire timeline.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":               map[string]any{"type": "integer", "description": "Report ID."},
				"name":             map[string]any{"type": "string"},
				"target_kind":      map[string]any{"type": "string"},
				"description":      map[string]any{"type": "string"},
				"briefing_content": map[string]any{"type": "string"},
				"timeline_entries": map[string]any{
					"type":        "array",
					"description": "Full replacement timeline. When provided, existing timeline rows are deleted and replaced by this exact list.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"datetime": map[string]any{"type": "string"},
							"title":    map[string]any{"type": "string"},
							"content":  map[string]any{"type": "string"},
						},
						"required": []string{"datetime", "title", "content"},
					},
				},
			},
			"required": []string{"id"},
		},
	}
}

func (editReportTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p editReportArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	id := p.ID
	if id == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var report Report
	if err := r.db.WithContext(ctx).First(&report, id).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_report load", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	existing, err := loadReportPageData(ctx, r.db.WithContext(ctx), id)
	if err != nil {
		slog.Error("lacerate: lookup tool edit_report load page data", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	changed := false
	name := report.Name
	desc := report.Description
	kind := report.Kind
	if p.Name != nil {
		s := strings.TrimSpace(*p.Name)
		if s == "" {
			err := fmt.Errorf("name cannot be empty")
			slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		name = s
		changed = true
	}
	if p.TargetKind != nil {
		s := strings.TrimSpace(*p.TargetKind)
		if !validReportKind(s) {
			err := fmt.Errorf("invalid target_kind")
			slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		kind = s
		changed = true
	}
	if p.Description != nil {
		desc = strings.TrimSpace(*p.Description)
		changed = true
	}
	formData, err := reportFormDataFromExistingToolArgs(ctx, existing, kind, p.BriefingContent, p.TimelineEntries)
	if err != nil {
		slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.BriefingContent != nil || p.TimelineEntries != nil || kind != report.Kind {
		changed = true
	}
	if !changed {
		err := fmt.Errorf("no fields to update")
		slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	formData.Name = name
	formData.Description = desc
	if err := saveReportToolUpdate(ctx, r.db, id, name, desc, kind, formData); err != nil {
		slog.Error("lacerate: lookup tool edit_report save", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": id, "name": name, "kind": kind}, nil
}

type appendTimelineEntriesTool struct{}

func (appendTimelineEntriesTool) Name() string { return "append_timeline_entries" }

func (appendTimelineEntriesTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "append_timeline_entries",
		Description: "Append new entries to an existing timeline report by ID. Use this when you want to keep the current timeline intact and add more events in chronological order.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer", "description": "Timeline report ID."},
				"timeline_entries": map[string]any{
					"type":        "array",
					"description": "New timeline entries to append. Existing entries are kept.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"datetime": map[string]any{"type": "string", "description": "Entry datetime in RFC3339."},
							"title":    map[string]any{"type": "string", "description": "Short entry title."},
							"content":  map[string]any{"type": "string", "description": "Markdown body for the entry."},
						},
						"required": []string{"datetime", "title", "content"},
					},
				},
			},
			"required": []string{"id", "timeline_entries"},
		},
	}
}

func (appendTimelineEntriesTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p appendTimelineEntriesArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.ID == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if len(p.TimelineEntries) == 0 {
		err := fmt.Errorf("timeline_entries is required")
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var report Report
	if err := r.db.WithContext(ctx).First(&report, p.ID).Error; err != nil {
		slog.Error("lacerate: lookup tool append_timeline_entries load", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if report.Kind != "timeline" {
		err := fmt.Errorf("report %d is %q, not timeline", p.ID, report.Kind)
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	existing, err := loadReportPageData(ctx, r.db.WithContext(ctx), p.ID)
	if err != nil {
		slog.Error("lacerate: lookup tool append_timeline_entries load page data", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	currentEntries, err := parsedTimelineEntriesFromExisting(existing)
	if err != nil {
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	newEntries, err := parseToolTimelineEntries(ctx, p.TimelineEntries)
	if err != nil {
		slog.Warn("lacerate: lookup tool append_timeline_entries", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	formData := parsedReportFormData{
		Name:            report.Name,
		Description:     report.Description,
		Kind:            "timeline",
		TimelineEntries: append(currentEntries, newEntries...),
	}
	if err := saveReportToolUpdate(ctx, r.db, p.ID, report.Name, report.Description, report.Kind, formData); err != nil {
		slog.Error("lacerate: lookup tool append_timeline_entries save", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{
		"id":             p.ID,
		"name":           report.Name,
		"kind":           report.Kind,
		"appended_count": len(newEntries),
		"total_entries":  len(formData.TimelineEntries),
	}, nil
}

func saveReportToolUpdate(ctx context.Context, db *gorm.DB, id uint, name, desc, kind string, formData parsedReportFormData) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Report{Model: gorm.Model{ID: id}}).Updates(map[string]any{
			"Name":        name,
			"Description": desc,
			"Kind":        kind,
		}).Error; err != nil {
			return err
		}
		if err := deleteReportKindExtensionRows(tx, id); err != nil {
			return err
		}
		if err := createReportKindRow(tx, id, formData); err != nil {
			return err
		}
		return refreshReportEmbedding(ctx, tx, id)
	})
}

type getRelevantReportsTool struct{}

func (getRelevantReportsTool) Name() string { return "get_relevant_reports" }

func (getRelevantReportsTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "get_relevant_reports",
		Description: "Ranked cosine similarity search over reports with embeddings.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Natural language query text."},
				"limit": map[string]any{"type": "integer", "description": "Max rows (default 10)."},
			},
			"required": []string{"query"},
		},
	}
}

func (getRelevantReportsTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p embeddingSearchArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_reports", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	q := strings.TrimSpace(p.Query)
	if q == "" {
		err := fmt.Errorf("query is required")
		slog.Warn("lacerate: lookup tool get_relevant_reports", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	limit, err := parseLookupSearchLimit(p.Limit, 10, 50)
	if err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_reports", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	e := vlEmbedder()
	if e == nil {
		err := fmt.Errorf("embedding service not configured")
		slog.Warn("lacerate: lookup tool get_relevant_reports", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	vec, err := e.Embed(ctx, q)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_reports embed", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if len(vec) != IntelEmbeddingDim {
		err := fmt.Errorf("embedding dimension %d, want %d", len(vec), IntelEmbeddingDim)
		slog.Error("lacerate: lookup tool get_relevant_reports", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	reports, err := searchReportsByEmbedding(r.db.WithContext(ctx), pgvector.NewVector(vec), limit)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_reports search", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	pageData, err := loadReportPageDataList(ctx, r.db.WithContext(ctx), reports)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_reports page data", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	out := make([]map[string]any, 0, len(pageData))
	for _, data := range pageData {
		snippet := reportPageDataSnippet(data)
		if len(snippet) > 400 {
			snippet = snippet[:397] + "..."
		}
		out = append(out, map[string]any{
			"id": data.Report.ID, "name": data.Report.Name, "kind": data.Report.Kind, "snippet": snippet,
		})
	}
	return map[string]any{"reports": out}, nil
}

type getRelevantIntelTool struct{}

func (getRelevantIntelTool) Name() string { return "get_relevant_intel" }

func (getRelevantIntelTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "get_relevant_intel",
		Description: "Ranked cosine similarity search over ingested intel records.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Natural language query text."},
				"limit": map[string]any{"type": "integer", "description": "Max rows (default 10)."},
			},
			"required": []string{"query"},
		},
	}
}

func (getRelevantIntelTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p embeddingSearchArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	q := strings.TrimSpace(p.Query)
	if q == "" {
		err := fmt.Errorf("query is required")
		slog.Warn("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	limit, err := parseLookupSearchLimit(p.Limit, 10, 50)
	if err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	e := vlEmbedder()
	if e == nil {
		err := fmt.Errorf("embedding service not configured")
		slog.Warn("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	vec, err := e.Embed(ctx, q)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_intel embed", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if len(vec) != IntelEmbeddingDim {
		err := fmt.Errorf("embedding dimension %d, want %d", len(vec), IntelEmbeddingDim)
		slog.Error("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	intels, err := searchIntelByEmbedding(r.db.WithContext(ctx), pgvector.NewVector(vec), limit)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_intel search", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	out := make([]map[string]any, 0, len(intels))
	for _, i := range intels {
		c := strings.TrimSpace(i.Content)
		if len(c) > 500 {
			c = c[:497] + "..."
		}
		out = append(out, map[string]any{
			"id":        i.ID,
			"source_id": i.SourceID,
			"datetime":  i.Datetime.UTC().Format(time.RFC3339),
			"snippet":   c,
		})
	}
	return map[string]any{"intel": out}, nil
}

type createTargetOfInterestTool struct{}

func (createTargetOfInterestTool) Name() string { return "create_target_of_interest" }

func (createTargetOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "create_target_of_interest",
		Description: "Create a target of interest (TOI): a short, accurate summary of an entity the user tracks. Use sparingly—only when a new entity needs tracking or context warrants a new row.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string", "description": "Distinct entity label."},
				"description": map[string]any{"type": "string", "description": "Brief factual description."},
			},
			"required": []string{"name"},
		},
	}
}

func (createTargetOfInterestTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p createTargetOfInterestArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool create_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		err := fmt.Errorf("name is required")
		slog.Warn("lacerate: lookup tool create_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	desc := ""
	if p.Description != nil {
		desc = strings.TrimSpace(*p.Description)
	}
	t := TargetOfInterest{Name: name, Description: desc}
	if err := r.db.WithContext(ctx).Create(&t).Error; err != nil {
		slog.Error("lacerate: lookup tool create_target_of_interest db", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": t.ID, "name": t.Name}, nil
}

type editTargetOfInterestTool struct{}

func (editTargetOfInterestTool) Name() string { return "edit_target_of_interest" }

func (editTargetOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "edit_target_of_interest",
		Description: "Update an existing target of interest by ID. Use only when new information materially changes the entity's context. Only include fields to change.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "integer", "description": "Target of interest ID."},
				"name":        map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
			},
			"required": []string{"id"},
		},
	}
}

func (editTargetOfInterestTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p editTargetOfInterestArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if p.ID == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var t TargetOfInterest
	if err := r.db.WithContext(ctx).First(&t, p.ID).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_target_of_interest load", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	changed := false
	if p.Name != nil {
		s := strings.TrimSpace(*p.Name)
		if s == "" {
			err := fmt.Errorf("name cannot be empty")
			slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		t.Name = s
		changed = true
	}
	if p.Description != nil {
		t.Description = strings.TrimSpace(*p.Description)
		changed = true
	}
	if !changed {
		err := fmt.Errorf("no fields to update")
		slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if err := r.db.WithContext(ctx).Save(&t).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_target_of_interest save", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": t.ID, "name": t.Name}, nil
}

type getRelevantTargetsOfInterestTool struct{}

func (getRelevantTargetsOfInterestTool) Name() string { return "get_relevant_targets_of_interest" }

func (getRelevantTargetsOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "get_relevant_targets_of_interest",
		Description: "Ranked cosine similarity search over targets of interest with embeddings. Use natural-language query text to find related entities.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Natural language query text."},
				"limit": map[string]any{"type": "integer", "description": "Max rows (default 10)."},
			},
			"required": []string{"query"},
		},
	}
}

func (getRelevantTargetsOfInterestTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	var p embeddingSearchArgs
	if err := unmarshalToolArgs(args, &p); err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	q := strings.TrimSpace(p.Query)
	if q == "" {
		err := fmt.Errorf("query is required")
		slog.Warn("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	limit, err := parseLookupSearchLimit(p.Limit, 10, 50)
	if err != nil {
		slog.Warn("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	e := vlEmbedder()
	if e == nil {
		err := fmt.Errorf("embedding service not configured")
		slog.Warn("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	vec, err := e.Embed(ctx, q)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_targets_of_interest embed", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if len(vec) != IntelEmbeddingDim {
		err := fmt.Errorf("embedding dimension %d, want %d", len(vec), IntelEmbeddingDim)
		slog.Error("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	rows, err := searchTargetsOfInterestByEmbedding(r.db.WithContext(ctx), pgvector.NewVector(vec), limit)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_targets_of_interest search", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		snippet := strings.TrimSpace(row.Description)
		if len(snippet) > 400 {
			snippet = snippet[:397] + "..."
		}
		out = append(out, map[string]any{"id": row.ID, "name": row.Name, "snippet": snippet})
	}
	return map[string]any{"targets_of_interest": out}, nil
}

func validReportKind(k string) bool {
	k = strings.TrimSpace(k)
	if k == "" {
		return false
	}
	for _, p := range ReportKindChoices {
		if p.Key == k {
			return true
		}
	}
	return false
}

func reportFormDataFromToolArgs(ctx context.Context, kind string, briefingContent *string, timelineEntries []reportTimelineEntryArgs) (parsedReportFormData, error) {
	out := parsedReportFormData{Kind: kind}
	switch kind {
	case "briefing":
		if briefingContent == nil || strings.TrimSpace(*briefingContent) == "" {
			return out, fmt.Errorf("briefing_content is required for briefing reports")
		}
		out.BriefingContent = strings.TrimSpace(*briefingContent)
	case "timeline":
		entries, err := parseToolTimelineEntries(ctx, timelineEntries)
		if err != nil {
			return out, err
		}
		out.TimelineEntries = entries
	default:
		return out, fmt.Errorf("unsupported report kind %q", kind)
	}
	return out, nil
}

func reportFormDataFromExistingToolArgs(ctx context.Context, existing ReportPageData, kind string, briefingContent *string, timelineEntries *[]reportTimelineEntryArgs) (parsedReportFormData, error) {
	out := parsedReportFormData{Kind: kind}
	switch kind {
	case "briefing":
		if briefingContent != nil {
			s := strings.TrimSpace(*briefingContent)
			if s == "" {
				return out, fmt.Errorf("briefing_content cannot be empty")
			}
			out.BriefingContent = s
			return out, nil
		}
		if existing.Briefing == nil {
			return out, fmt.Errorf("briefing_content is required when switching to briefing")
		}
		out.BriefingContent = existing.Briefing.Content
	case "timeline":
		if timelineEntries != nil {
			entries, err := parseToolTimelineEntries(ctx, *timelineEntries)
			if err != nil {
				return out, err
			}
			out.TimelineEntries = entries
			return out, nil
		}
		entries, err := parsedTimelineEntriesFromExisting(existing)
		if err != nil {
			return out, err
		}
		out.TimelineEntries = entries
	default:
		return out, fmt.Errorf("unsupported report kind %q", kind)
	}
	return out, nil
}

func parsedTimelineEntriesFromExisting(existing ReportPageData) ([]parsedReportTimelineEntry, error) {
	if existing.Timeline == nil {
		return nil, fmt.Errorf("timeline_entries are required when switching to timeline")
	}
	out := make([]parsedReportTimelineEntry, 0, len(existing.Timeline.Entries))
	for _, entry := range existing.Timeline.Entries {
		out = append(out, parsedReportTimelineEntry{
			Datetime: entry.Datetime,
			Title:    entry.Title,
			Content:  entry.Content,
		})
	}
	return out, nil
}

func parseToolTimelineEntries(ctx context.Context, rows []reportTimelineEntryArgs) ([]parsedReportTimelineEntry, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("timeline_entries are required for timeline reports")
	}
	tz, _ := ctx.Value("$tz").(*time.Location)
	if tz == nil {
		tz = time.UTC
	}
	out := make([]parsedReportTimelineEntry, 0, len(rows))
	for i, row := range rows {
		title := strings.TrimSpace(row.Title)
		content := strings.TrimSpace(row.Content)
		if title == "" {
			return nil, fmt.Errorf("timeline entry %d title is required", i+1)
		}
		if content == "" {
			return nil, fmt.Errorf("timeline entry %d content is required", i+1)
		}
		dtRaw := strings.TrimSpace(row.Datetime)
		if dtRaw == "" {
			return nil, fmt.Errorf("timeline entry %d datetime is required", i+1)
		}
		dt, err := time.Parse(time.RFC3339, dtRaw)
		if err != nil {
			dt, err = time.ParseInLocation("2006-01-02T15:04", dtRaw, tz)
			if err != nil {
				return nil, fmt.Errorf("timeline entry %d datetime must be RFC3339", i+1)
			}
		}
		out = append(out, parsedReportTimelineEntry{
			Datetime: dt,
			Title:    title,
			Content:  content,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Datetime.Before(out[j].Datetime)
	})
	return out, nil
}
