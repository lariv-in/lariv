package sqlagent

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"gorm.io/gorm"
)

const sqlToolName = "execute_sql"

type sqlAgentGormTxKey struct{}

// ContextWithGormTx attaches a *gorm.DB (typically a transaction from db.Transaction) so execute_sql runs on that handle.
func ContextWithGormTx(ctx context.Context, gdb *gorm.DB) context.Context {
	return context.WithValue(ctx, sqlAgentGormTxKey{}, gdb)
}

func gormTxFromContext(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(sqlAgentGormTxKey{})
	if v == nil {
		return nil, false
	}
	tx, ok := v.(*gorm.DB)
	return tx, ok && tx != nil
}

type sqlToolInput struct {
	SQL string `json:"sql"`
}

type sqlToolOutput struct {
	Result any `json:"result"`
}

func sqlLooksLikeRowReturning(sql string) bool {
	s := strings.TrimSpace(sql)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	for strings.HasPrefix(lower, "--") {
		if i := strings.IndexByte(lower, '\n'); i >= 0 {
			lower = strings.TrimSpace(lower[i+1:])
		} else {
			return false
		}
	}
	first := lower
	if i := strings.IndexAny(lower, " \t\n("); i > 0 {
		first = lower[:i]
	}
	switch first {
	case "select", "with", "show", "explain", "describe", "desc", "pragma":
		return true
	}
	return strings.Contains(lower, " returning ")
}

func sqlCellToJSONable(v any) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return x
	}
}

func runSQLQuery(tx *gorm.DB, sql string) (any, error) {
	rows, err := tx.Raw(sql).Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			slog.Error("sqlagent: close sql rows", "error", cerr)
		}
	}()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = sqlCellToJSONable(raw[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"columns": cols, "rows": out}, nil
}

func sqlToolHandler(tctx tool.Context, in sqlToolInput) (sqlToolOutput, error) {
	sqlStr := strings.TrimSpace(in.SQL)
	slog.Info("sqlagent: execute_sql", "sql", sqlStr)
	if sqlStr == "" {
		err := errors.New("sql is required")
		logError("sqlagent: execute_sql", err)
		return sqlToolOutput{}, err
	}
	tx, ok := gormTxFromContext(tctx)
	if !ok {
		err := errors.New("database transaction not available in tool context")
		logError("sqlagent: execute_sql", err)
		return sqlToolOutput{}, err
	}
	var result any
	var err error
	if sqlLooksLikeRowReturning(sqlStr) {
		result, err = runSQLQuery(tx, sqlStr)
	} else {
		r := tx.Exec(sqlStr)
		if r.Error != nil {
			logError("sqlagent: execute_sql exec", r.Error)
			return sqlToolOutput{}, r.Error
		}
		result = map[string]any{"rows_affected": r.RowsAffected}
	}
	if err != nil {
		logError("sqlagent: execute_sql query", err)
		return sqlToolOutput{}, err
	}
	return sqlToolOutput{Result: result}, nil
}

func newSQLTool() (tool.Tool, error) {
	t, err := functiontool.New(functiontool.Config{
		Name:        sqlToolName,
		Description: `Runs a single raw SQL statement on the app's database within the current request transaction. Field: "sql" (string). Statements that return rows (SELECT, WITH, SHOW, EXPLAIN, DESCRIBE, PRAGMA, or SQL containing RETURNING) yield {"columns":[...], "rows":[{col: value, ...}, ...]}. Other statements yield {"rows_affected": n}. Use standard SQL for the deployment's database (often PostgreSQL). Avoid destructive operations unless the user explicitly wants them.`,
	}, sqlToolHandler)
	if err != nil {
		logError("sqlagent: functiontool New execute_sql", err)
	}
	return t, err
}
