package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/registry"
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

const osintSpecialistSystem = `You are an OSINT Specialist: methodical, careful, and ethical. You only use the provided tools to gather and record intelligence from this application's corpora. Prefer verifiable facts from tool results over speculation. Before creating a new Target of Interest, check for an existing relevant report first; if one already exists for the same subject, update it with edit_target_of_interest instead of creating a duplicate. When done, give a concise summary of what you did and what you learned.`

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
					if name == "create_target_of_interest" || name == "edit_target_of_interest" {
						if targetID, ok := targetOfInterestIDFromToolOutput(out); ok && targetID > 0 {
							act := "create"
							if name == "edit_target_of_interest" {
								act = "edit"
							}
							if err := run.recordLookupTouchedTargetOfInterest(toolRow, act, targetID); err != nil {
								slog.Error("lacerate: lookup agent record touched Target of Interest", "error", err, "lookup_id", lu.ID, "target_of_interest_id", targetID)
							}
						}
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
		createTargetOfInterestTool{},
		editTargetOfInterestTool{},
		getRelevantTargetsOfInterestTool{},
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

func targetOfInterestIDFromToolOutput(out any) (uint, bool) {
	m, ok := out.(map[string]any)
	if !ok {
		return 0, false
	}
	id, ok := argUint(m, "id")
	return id, ok && id > 0
}

func (r *lookupRun) recordLookupTouchedTargetOfInterest(toolRow *LookupToolCall, action string, targetID uint) error {
	if r == nil || r.db == nil || toolRow == nil || toolRow.LookupLogEntryID == 0 || targetID == 0 {
		return nil
	}
	if action != "create" && action != "edit" {
		err := fmt.Errorf("invalid touch action %q", action)
		slog.Warn("lacerate: lookup touched Target of Interest", "error", err, "lookup_id", r.lookupID)
		return err
	}
	row := LookupLogTargetOfInterest{
		LookupID:           r.lookupID,
		LookupLogEntryID:   toolRow.LookupLogEntryID,
		TargetOfInterestID: targetID,
		Action:             action,
	}
	if err := r.db.Create(&row).Error; err != nil {
		slog.Error("lacerate: lookup touched Target of Interest insert", "error", err, "lookup_id", r.lookupID, "target_of_interest_id", targetID)
		return err
	}
	return nil
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
			if err := r.logThought(t); err != nil {
				slog.Error("lacerate: lookup log thought", "error", err, "lookup_id", r.lookupID)
			}
			continue
		}
		if err := r.logText(t); err != nil {
			slog.Error("lacerate: lookup log text", "error", err, "lookup_id", r.lookupID)
		}
	}
}

func (r *lookupRun) logThought(text string) error {
	entry := LookupLogEntry{LookupID: r.lookupID, Kind: "thought"}
	if err := r.db.Create(&entry).Error; err != nil {
		slog.Error("lacerate: lookup log thought entry", "error", err, "lookup_id", r.lookupID)
		return err
	}
	if err := r.db.Create(&LookupThought{LookupLogEntryID: entry.ID, Text: text}).Error; err != nil {
		slog.Error("lacerate: lookup log thought row", "error", err, "lookup_id", r.lookupID)
		return err
	}
	return nil
}

func (r *lookupRun) logText(text string) error {
	entry := LookupLogEntry{LookupID: r.lookupID, Kind: "text"}
	if err := r.db.Create(&entry).Error; err != nil {
		slog.Error("lacerate: lookup log text entry", "error", err, "lookup_id", r.lookupID)
		return err
	}
	if err := r.db.Create(&LookupText{LookupLogEntryID: entry.ID, Text: text}).Error; err != nil {
		slog.Error("lacerate: lookup log text row", "error", err, "lookup_id", r.lookupID)
		return err
	}
	return nil
}

func (r *lookupRun) startToolCall(name string, args map[string]any) (*LookupToolCall, error) {
	entry := LookupLogEntry{LookupID: r.lookupID, Kind: "tool_call"}
	if err := r.db.Create(&entry).Error; err != nil {
		slog.Error("lacerate: lookup tool call log entry", "error", err, "lookup_id", r.lookupID, "tool", name)
		return nil, err
	}
	argJSON, err := json.Marshal(args)
	if err != nil {
		slog.Error("lacerate: lookup tool call marshal args", "error", err, "lookup_id", r.lookupID, "tool", name)
		return nil, err
	}
	tool := &LookupToolCall{
		LookupLogEntryID: entry.ID,
		Name:             name,
		Arguments:        datatypes.JSON(argJSON),
	}
	if err := r.db.Create(tool).Error; err != nil {
		slog.Error("lacerate: lookup tool call row", "error", err, "lookup_id", r.lookupID, "tool", name)
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
	entry := LookupLogEntry{LookupID: r.lookupID, Kind: "tool_error"}
	if err := r.db.Create(&entry).Error; err != nil {
		slog.Error("lacerate: lookup tool error entry", "error", err, "lookup_id", r.lookupID)
		return err
	}
	var det datatypes.JSON
	if detail != nil {
		b, err := json.Marshal(detail)
		if err != nil {
			slog.Error("lacerate: lookup tool error marshal detail", "error", err, "lookup_id", r.lookupID)
			return err
		}
		det = b
	}
	if err := r.db.Create(&LookupToolError{
		LookupLogEntryID: entry.ID,
		ToolName:         toolName,
		Message:          message,
		Detail:           det,
	}).Error; err != nil {
		slog.Error("lacerate: lookup tool error row", "error", err, "lookup_id", r.lookupID)
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

type createTargetOfInterestTool struct{}

func (createTargetOfInterestTool) Name() string { return "create_target_of_interest" }

func (createTargetOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "create_target_of_interest",
		Description: "Create a curated Target of Interest (report, briefing, etc.) in the database. Before creating a report, check for an existing relevant report and use edit_target_of_interest instead if one already exists.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string", "description": "Short title for the Target of Interest."},
				"target_type": map[string]any{"type": "string", "description": "One of: report, briefing, memo, dataset, other."},
				"description": map[string]any{"type": "string", "description": "Summary or context."},
				"content":     map[string]any{"type": "string", "description": "Body text / markdown for the Target of Interest."},
			},
			"required": []string{"name", "target_type", "content"},
		},
	}
}

func (createTargetOfInterestTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	name := strings.TrimSpace(argString(args, "name"))
	typ := strings.TrimSpace(argString(args, "asset_type"))
	if typ == "" {
		typ = strings.TrimSpace(argString(args, "type"))
	}
	desc := strings.TrimSpace(argString(args, "description"))
	content := strings.TrimSpace(argString(args, "content"))
	if name == "" {
		err := fmt.Errorf("name is required")
		slog.Warn("lacerate: lookup tool create_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if content == "" {
		err := fmt.Errorf("content is required")
		slog.Warn("lacerate: lookup tool create_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if !validTargetOfInterestType(typ) {
		err := fmt.Errorf("target_type must be one of the allowed keys (report, briefing, memo, dataset, other)")
		slog.Warn("lacerate: lookup tool create_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	a := TargetOfInterest{Name: name, Type: typ, Description: desc, Content: content}
	if err := r.db.WithContext(ctx).Create(&a).Error; err != nil {
		slog.Error("lacerate: lookup tool create_target_of_interest db", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": a.ID, "name": a.Name, "type": a.Type}, nil
}

type editTargetOfInterestTool struct{}

func (editTargetOfInterestTool) Name() string { return "edit_target_of_interest" }

func (editTargetOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "edit_target_of_interest",
		Description: "Update an existing Target of Interest by ID. Use this when you found an existing report for the same subject and want to revise it instead of creating a duplicate. Only include fields you want to change.",
		ParametersJsonSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "integer", "description": "Target of Interest ID."},
				"name":        map[string]any{"type": "string"},
				"target_type": map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
				"content":     map[string]any{"type": "string"},
			},
			"required": []string{"id"},
		},
	}
}

func (editTargetOfInterestTool) Run(ctx context.Context, r *lookupRun, args map[string]any) (any, error) {
	id, ok := argUint(args, "id")
	if !ok || id == 0 {
		err := fmt.Errorf("id is required")
		slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	var a TargetOfInterest
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_target_of_interest load", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	changed := false
	if s, ok := argOptionalString(args, "name"); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			err := fmt.Errorf("name cannot be empty")
			slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Name = s
		changed = true
	}
	if s, ok := argOptionalString(args, "target_type"); ok {
		s = strings.TrimSpace(s)
		if !validTargetOfInterestType(s) {
			err := fmt.Errorf("invalid target_type")
			slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Type = s
		changed = true
	} else if s, ok := argOptionalString(args, "asset_type"); ok {
		s = strings.TrimSpace(s)
		if !validTargetOfInterestType(s) {
			err := fmt.Errorf("invalid asset_type")
			slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Type = s
		changed = true
	}
	if s, ok := argOptionalString(args, "type"); ok {
		s = strings.TrimSpace(s)
		if !validTargetOfInterestType(s) {
			err := fmt.Errorf("invalid type")
			slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
			return nil, err
		}
		a.Type = s
		changed = true
	}
	if s, ok := argOptionalString(args, "description"); ok {
		a.Description = s
		changed = true
	}
	if s, ok := argOptionalString(args, "content"); ok {
		a.Content = s
		changed = true
	}
	if !changed {
		err := fmt.Errorf("no fields to update")
		slog.Warn("lacerate: lookup tool edit_target_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	if err := r.db.WithContext(ctx).Save(&a).Error; err != nil {
		slog.Error("lacerate: lookup tool edit_target_of_interest save", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	return map[string]any{"id": a.ID, "name": a.Name, "type": a.Type}, nil
}

type getRelevantTargetsOfInterestTool struct{}

func (getRelevantTargetsOfInterestTool) Name() string { return "get_relevant_targets_of_interest" }

func (getRelevantTargetsOfInterestTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "get_relevant_targets_of_interest",
		Description: "Ranked cosine similarity search over Targets of Interest with embeddings.",
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
	q := strings.TrimSpace(argString(args, "query"))
	if q == "" {
		err := fmt.Errorf("query is required")
		slog.Warn("lacerate: lookup tool get_relevant_targets_of_interest", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	limit := clampLookupAgentLimit(int(argNumber(args, "limit", 10)))
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
	targets, err := searchTargetsOfInterestByEmbedding(r.db.WithContext(ctx), vec, limit)
	if err != nil {
		slog.Error("lacerate: lookup tool get_relevant_targets_of_interest search", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	out := make([]map[string]any, 0, len(targets))
	for _, t := range targets {
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
	return map[string]any{"targets_of_interest": out}, nil
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
	q := strings.TrimSpace(argString(args, "query"))
	if q == "" {
		err := fmt.Errorf("query is required")
		slog.Warn("lacerate: lookup tool get_relevant_intel", "error", err, "lookup_id", r.lookupID)
		return nil, err
	}
	limit := clampLookupAgentLimit(int(argNumber(args, "limit", 10)))
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
	intels, err := searchIntelByEmbedding(r.db.WithContext(ctx), vec, limit)
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

func validTargetOfInterestType(k string) bool {
	k = strings.TrimSpace(k)
	if k == "" {
		return false
	}
	for _, p := range TargetOfInterestTypeChoices {
		if p.Key == k {
			return true
		}
	}
	return false
}

func argString(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func argOptionalString(m map[string]any, k string) (string, bool) {
	v, ok := m[k]
	if !ok || v == nil {
		return "", false
	}
	s := fmt.Sprint(v)
	return s, true
}

func argUint(m map[string]any, k string) (uint, bool) {
	v, ok := m[k]
	if !ok || v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		if x < 0 || x > float64(^uint(0)>>1) {
			return 0, false
		}
		return uint(x), true
	case int:
		if x < 0 {
			return 0, false
		}
		return uint(x), true
	case int64:
		if x < 0 {
			return 0, false
		}
		return uint(x), true
	case json.Number:
		n, err := strconv.ParseUint(string(x), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	case string:
		n, err := strconv.ParseUint(strings.TrimSpace(x), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	default:
		n, err := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(v)), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	}
}

func argNumber(m map[string]any, k string, def float64) float64 {
	v, ok := m[k]
	if !ok || v == nil {
		return def
	}
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		if err != nil {
			return def
		}
		return f
	default:
		f, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(v)), 64)
		if err != nil {
			return def
		}
		return f
	}
}

func clampLookupAgentLimit(n int) int {
	if n < 1 {
		return 10
	}
	if n > 50 {
		return 50
	}
	return n
}
