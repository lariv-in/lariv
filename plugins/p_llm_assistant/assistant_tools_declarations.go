package p_llm_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/lariv-in/lago/registry"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

// LlmTool defines the interface for dynamically registered LLM agent tools.
type LlmTool interface {
	Name() string
	Declaration() *genai.FunctionDeclaration
	Run(ctx context.Context, db *gorm.DB, args map[string]any) (res map[string]any, err error)
}

// LlmToolRegistry is the global registry where plugins can register their LLM tools.
var LlmToolRegistry = registry.NewRegistry[LlmTool]()

type googleSearchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type googleSearchTool struct{}

func (t *googleSearchTool) Name() string {
	return "google_search"
}

func (t *googleSearchTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "google_search",
		Description: "Search the public web via Google Custom Search (configured in Lago). Use when you need to search or verify details on the web.",
		Parameters:  p_google_genai.NewSchema[googleSearchArgs](),
	}
}

func (t *googleSearchTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (res map[string]any, err error) {
	var sArgs googleSearchArgs
	if b, jerr := json.Marshal(args); jerr == nil {
		_ = json.Unmarshal(b, &sArgs)
	}
	hits, err := runGoogleSearchTool(ctx, sArgs.Query, sArgs.Limit)
	if err != nil {
		return nil, err
	}
	return map[string]any{"hits": hits}, nil
}

type listSkillsArgs struct{}

type listSkillsTool struct{}

type skillListResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (t *listSkillsTool) Name() string {
	return "list_skills"
}

func (t *listSkillsTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "list_skills",
		Description: "Retrieve a list of all assistant skills, including their names and descriptions.",
		Parameters:  p_google_genai.NewSchema[listSkillsArgs](),
	}
}

func (t *listSkillsTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (res map[string]any, err error) {
	var skills []Skill
	if err := db.WithContext(ctx).Order("name ASC").Find(&skills).Error; err != nil {
		return nil, err
	}
	out := make([]skillListResponse, 0, len(skills))
	for _, s := range skills {
		out = append(out, skillListResponse{
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return map[string]any{"skills": out}, nil
}

type getSkillDetailArgs struct {
	Name string `json:"name"`
}

type getSkillDetailTool struct{}

func (t *getSkillDetailTool) Name() string {
	return "get_skill_detail"
}

func (t *getSkillDetailTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "get_skill_detail",
		Description: "Retrieve detailed description of a skill by name, including its content and associated file paths.",
		Parameters:  p_google_genai.NewSchema[getSkillDetailArgs](),
	}
}

func (t *getSkillDetailTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (res map[string]any, err error) {
	var sArgs getSkillDetailArgs
	if b, jerr := json.Marshal(args); jerr == nil {
		_ = json.Unmarshal(b, &sArgs)
	}
	if strings.TrimSpace(sArgs.Name) == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	var skill Skill
	if err := db.WithContext(ctx).Preload("Files").Where("name = ?", sArgs.Name).First(&skill).Error; err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(skill.Files))
	for _, f := range skill.Files {
		paths = append(paths, f.GetPath(db))
	}

	out := map[string]any{
		"name":        skill.Name,
		"description": skill.Description,
		"content":     skill.Content,
		"file_paths":  paths,
	}
	return out, nil
}

type readFileArgs struct {
	Path string `json:"path"`
}

type readFileTool struct{}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "read_file",
		Description: "Read the content of a virtual file VNode using its file path (e.g., /Skills/code.py).",
		Parameters:  p_google_genai.NewSchema[readFileArgs](),
	}
}

func (t *readFileTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (res map[string]any, err error) {
	var rArgs readFileArgs
	if b, jerr := json.Marshal(args); jerr == nil {
		_ = json.Unmarshal(b, &rArgs)
	}
	if strings.TrimSpace(rArgs.Path) == "" {
		return nil, fmt.Errorf("file path is required")
	}

	node, _, nerr := p_filesystem.GetVNodeByPath(db, rArgs.Path)
	if nerr != nil {
		return nil, nerr
	}
	if node == nil {
		return nil, fmt.Errorf("file not found at path %q", rArgs.Path)
	}
	if node.IsDirectory {
		return nil, fmt.Errorf("path %q is a directory, not a file", rArgs.Path)
	}

	dl, dlerr := node.OpenDownload()
	if dlerr != nil {
		return nil, dlerr
	}
	defer dl.Reader.Close()

	contentBytes, readErr := io.ReadAll(dl.Reader)
	if readErr != nil {
		return nil, readErr
	}

	return map[string]any{"content": string(contentBytes)}, nil
}

func init() {
	LlmToolRegistry.Register("google_search", &googleSearchTool{})
	LlmToolRegistry.Register("list_skills", &listSkillsTool{})
	LlmToolRegistry.Register("get_skill_detail", &getSkillDetailTool{})
	LlmToolRegistry.Register("read_file", &readFileTool{})
}

// assistantGeminiTools returns function declarations loaded dynamically from [LlmToolRegistry].
func assistantGeminiTools() []*genai.Tool {
	var decls []*genai.FunctionDeclaration
	for _, tool := range LlmToolRegistry.All() {
		decls = append(decls, tool.Declaration())
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}
}
