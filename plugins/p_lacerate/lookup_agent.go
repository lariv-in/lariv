package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

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

const osintSpecialistSystem = `You are an OSINT Specialist: methodical, careful, and ethical. You only use the provided tools to gather and record intelligence from this application's corpora. Prefer verifiable facts from tool results over speculation. Before creating a new report, search for an existing one for the same subject; if one already exists, update it with edit_report instead of creating a duplicate. When done, give a concise summary of what you did and what you learned.`

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
		getRelevantReportsTool{},
		getRelevantIntelTool{},
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
		Description: "Create a curated report (briefing, memo, etc.) in the database. Before creating, check for an existing relevant report and use edit_report instead if one already exists.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string", "description": "Short title for the report."},
				"target_type": map[string]any{"type": "string", "description": "One of: report, briefing, memo, dataset, other."},
				"description": map[string]any{"type": "string", "description": "Summary or context."},
				"content":     map[string]any{"type": "string", "description": "Body text / markdown for the report."},
			},
			"required": []string{"name", "target_type", "content"},
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
	typ := strings.TrimSpace(p.TargetType)
	desc := ""
	if p.Description != nil {
		desc = strings.TrimSpace(*p.Description)
	}
	content := strings.TrimSpace(p.Content)
	if name == "" {
		err := fmt.Errorf("name is required")
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if content == "" {
		err := fmt.Errorf("content is required")
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if !validReportType(typ) {
		err := fmt.Errorf("target_type must be one of the allowed keys (report, briefing, memo, dataset, other)")
		slog.Warn("lacerate: lookup tool create_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	a := Report{Name: name, Type: typ, Description: desc, Content: content}
	if err := r.db.WithContext(ctx).Create(&a).Error; err != nil {
		slog.Error("lacerate: lookup tool create_report db", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": a.ID, "name": a.Name, "type": a.Type}, nil
}

type editReportTool struct{}

func (editReportTool) Name() string { return "edit_report" }

func (editReportTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "edit_report",
		Description: "Update an existing report by ID. Use when an existing report covers the same subject and should be revised instead of creating a duplicate. Only include fields you want to change.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "integer", "description": "Report ID."},
				"name":        map[string]any{"type": "string"},
				"target_type": map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
				"content":     map[string]any{"type": "string"},
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
	var a Report
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_report load", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	changed := false
	if p.Name != nil {
		s := strings.TrimSpace(*p.Name)
		if s == "" {
			err := fmt.Errorf("name cannot be empty")
			slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Name = s
		changed = true
	}
	if p.TargetType != nil {
		s := strings.TrimSpace(*p.TargetType)
		if !validReportType(s) {
			err := fmt.Errorf("invalid target_type")
			slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Type = s
		changed = true
	}
	if p.Description != nil {
		a.Description = *p.Description
		changed = true
	}
	if p.Content != nil {
		a.Content = *p.Content
		changed = true
	}
	if !changed {
		err := fmt.Errorf("no fields to update")
		slog.Warn("lacerate: lookup tool edit_report", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if err := r.db.WithContext(ctx).Save(&a).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_report save", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": a.ID, "name": a.Name, "type": a.Type}, nil
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
	out := make([]map[string]any, 0, len(reports))
	for _, t := range reports {
		snippet := strings.TrimSpace(t.Description)
		if snippet == "" {
			snippet = strings.TrimSpace(t.Content)
		}
		if len(snippet) > 400 {
			snippet = snippet[:397] + "..."
		}
		out = append(out, map[string]any{
			"id": t.ID, "name": t.Name, "type": t.Type, "snippet": snippet,
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
		out = append(out, map[string]any{"id": i.ID, "source_id": i.SourceID, "snippet": c})
	}
	return map[string]any{"intel": out}, nil
}

func validReportType(k string) bool {
	k = strings.TrimSpace(k)
	if k == "" {
		return false
	}
	for _, p := range ReportTypeChoices {
		if p.Key == k {
			return true
		}
	}
	return false
}
