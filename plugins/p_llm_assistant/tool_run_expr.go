package p_llm_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/builtin"
	"github.com/lariv-in/lariv/plugins/p_filesystem"
	"github.com/lariv-in/lariv/plugins/p_google_genai"
	"github.com/lariv-in/lariv/registry"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

// ExprEnvRegistry is the global registry that drives the environment exposed to
// expr expressions. Each entry is a named value (any type: scalar, func, struct)
// that becomes a top-level variable inside the expression.
//
// Other packages (plugins, deployments) register their contributions at init
// time, for example:
//
//	func init() {
//	    p_llm_assistant.ExprEnvRegistry.Register("now", time.Now)
//	}
var ExprEnvRegistry = registry.NewRegistry[any]()

// ContextualFunc allows registering functions or values that need the current
// request context or database connection.
type ContextualFunc func(ctx context.Context, db *gorm.DB) any

// exprEnv builds a map[string]any from ExprEnvRegistry suitable for passing to
// expr as the environment, resolving any ContextualFuncs.
func exprEnv(ctx context.Context, db *gorm.DB) map[string]any {
	all := ExprEnvRegistry.All()
	env := make(map[string]any, len(all))
	for k, v := range all {
		if cf, ok := v.(ContextualFunc); ok {
			env[k] = cf(ctx, db)
		} else {
			env[k] = v
		}
	}
	return env
}

// ---- run_expr tool ----

type runExprArgs struct {
	Expression string `json:"expression"`
}

type runExprTool struct{}

func (t *runExprTool) Name() string { return "run_expr" }

func (t *runExprTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "run_expr",
		Description: "Evaluate an expression using the expr-lang evaluator (https://github.com/expr-lang/expr). " +
			"Use this as a calculator for arithmetic (e.g. \"(100 + 200) * 1.18\"), unit conversions, " +
			"or any other numeric computation. Also supports data lookups and logic against a " +
			"pre-configured environment of variables and helper functions. " +
			"Returns the result as a JSON-encoded value, or an error string if evaluation fails.",
		Parameters: p_google_genai.NewSchema[runExprArgs](),
	}
}

func (t *runExprTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a runExprArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	if a.Expression == "" {
		return nil, fmt.Errorf("expression is required")
	}

	env := exprEnv(ctx, db)

	program, err := expr.Compile(a.Expression, expr.Env(env))
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	// JSON-encode result so any type (slice, map, struct) round-trips cleanly.
	encoded, err := json.Marshal(result)
	if err != nil {
		return map[string]any{"result": fmt.Sprint(result)}, nil
	}
	return map[string]any{"result": string(encoded)}, nil
}

// ---- run_expr_file tool ----

type runExprFileArgs struct {
	Path string         `json:"path"`
	Args map[string]any `json:"args,omitempty"`
}

type runExprFileTool struct{}

func (t *runExprFileTool) Name() string { return "run_expr_file" }

func (t *runExprFileTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name: "run_expr_file",
		Description: "Read and evaluate an expression script stored in a virtual file (VNode) using the expr-lang evaluator. " +
			"Merges the provided args into the common environment before running.",
		Parameters: p_google_genai.NewSchema[runExprFileArgs](),
	}
}

func (t *runExprFileTool) Run(ctx context.Context, db *gorm.DB, args map[string]any) (map[string]any, error) {
	var a runExprFileArgs
	if b, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	if a.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	node, _, nerr := p_filesystem.GetVNodeByPath(db, a.Path)
	if nerr != nil {
		return nil, nerr
	}
	if node == nil {
		return nil, fmt.Errorf("file not found at path %q", a.Path)
	}
	if node.IsDirectory {
		return nil, fmt.Errorf("path %q is a directory, not a file", a.Path)
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

	env := exprEnv(ctx, db)
	for k, v := range a.Args {
		env[k] = v
	}

	program, err := expr.Compile(string(contentBytes), expr.Env(env))
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return map[string]any{"error": err.Error()}, nil
	}

	// JSON-encode result so any type (slice, map, struct) round-trips cleanly.
	encoded, err := json.Marshal(result)
	if err != nil {
		return map[string]any{"result": fmt.Sprint(result)}, nil
	}
	return map[string]any{"result": string(encoded)}, nil
}

func init() {
	LlmToolRegistry.Register("run_expr", &runExprTool{})
	LlmToolRegistry.Register("run_expr_file", &runExprFileTool{})
	LlmToolRegistry.Register("list_expr_env", &listExprEnvTool{})
}

// ---- list_expr_env tool ----

type listExprEnvArgs struct{}

type listExprEnvTool struct{}

func (t *listExprEnvTool) Name() string { return "list_expr_env" }

func (t *listExprEnvTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "list_expr_env",
		Description: "List all variable and function names available in the expr environment. Use this to discover what is available before writing a run_expr expression.",
		Parameters:  p_google_genai.NewSchema[listExprEnvArgs](),
	}
}

func (t *listExprEnvTool) Run(_ context.Context, _ *gorm.DB, _ map[string]any) (map[string]any, error) {
	all := ExprEnvRegistry.All()
	envKeys := make([]string, 0, len(all))
	for k := range all {
		envKeys = append(envKeys, k)
	}
	return map[string]any{
		"env_variables":     envKeys,
		"builtin_functions": builtin.Names,
	}, nil
}
