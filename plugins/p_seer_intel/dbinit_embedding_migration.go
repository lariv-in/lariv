package p_seer_intel

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit("p_seer_intel.embedding_migration", func(db *gorm.DB) *gorm.DB {
		if db == nil || db.Name() != "postgres" {
			return db
		}
		ctx := context.Background()
		if err := migrateIntelEmbeddingColumn(ctx, db); err != nil {
			slog.Error("p_seer_intel: migrate embedding column", "error", err)
			return db
		}
		if err := backfillIntelEmbeddings(ctx, db); err != nil {
			slog.Warn("p_seer_intel: backfill embeddings skipped", "error", err)
		}
		return db
	})
}

func migrateIntelEmbeddingColumn(ctx context.Context, db *gorm.DB) error {
	tableName, err := intelTableName(db)
	if err != nil {
		return err
	}
	var currentType string
	err = db.WithContext(ctx).Raw(`
		SELECT format_type(a.atttypid, a.atttypmod)
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		WHERE c.relname = ? AND a.attname = 'embedding' AND a.attnum > 0 AND NOT a.attisdropped
		LIMIT 1
	`, tableName).Scan(&currentType).Error
	if err != nil {
		logIfAbortedTransaction(err, "scan embedding column type", "table", tableName)
		return err
	}
	wantType := fmt.Sprintf("vector(%d)", SeerIntelEmbeddingDim)
	if strings.EqualFold(strings.TrimSpace(currentType), wantType) {
		return nil
	}
	sql := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN embedding TYPE %s USING NULL`, tableName, wantType)
	if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
		logIfAbortedTransaction(err, "alter embedding column type", "table", tableName)
		return err
	}
	slog.Info("p_seer_intel: embedding column migrated", "table", tableName, "from", currentType, "to", wantType)
	return nil
}

func backfillIntelEmbeddings(ctx context.Context, db *gorm.DB) error {
	dim, err := p_google_genai.EmbeddingDimension(ctx)
	if err != nil {
		return err
	}
	if dim != SeerIntelEmbeddingDim {
		return fmt.Errorf("embedding model dimension %d does not match schema %d", dim, SeerIntelEmbeddingDim)
	}

	var rows []Intel
	if err := db.WithContext(ctx).
		Where("embedding IS NULL").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		logIfAbortedTransaction(err, "list null embeddings")
		return err
	}
	for i := range rows {
		row := rows[i]
		k, err := LoadIntelKind(ctx, db, strings.TrimSpace(row.Kind), row.KindID)
		if err != nil {
			slog.Warn("p_seer_intel: load intel kind for backfill", "intel_id", row.ID, "error", err)
			continue
		}
		if k == nil {
			slog.Warn("p_seer_intel: nil intel kind for backfill", "intel_id", row.ID)
			continue
		}
		content := strings.TrimSpace(k.Content())
		if content == "" {
			slog.Warn("p_seer_intel: empty content for backfill", "intel_id", row.ID)
			continue
		}
		values, err := p_google_genai.EmbedText(ctx, p_google_genai.EmbedTaskSearchDocument, content)
		if err != nil {
			slog.Warn("p_seer_intel: embed backfill", "intel_id", row.ID, "error", err)
			continue
		}
		if len(values) != SeerIntelEmbeddingDim {
			slog.Warn("p_seer_intel: backfill dimension mismatch", "intel_id", row.ID, "got", len(values), "want", SeerIntelEmbeddingDim)
			continue
		}
		vec := pgvector.NewVector(values)
		if err := db.WithContext(ctx).Model(&Intel{}).Where("id = ?", row.ID).Update("embedding", &vec).Error; err != nil {
			logIfAbortedTransaction(err, "save backfill embedding", "intel_id", row.ID)
			slog.Warn("p_seer_intel: save backfill embedding", "intel_id", row.ID, "error", err)
			continue
		}
	}
	return nil
}

func logIfAbortedTransaction(err error, operation string, attrs ...any) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "sqlstate 25p02") && !strings.Contains(msg, "current transaction is aborted") {
		return
	}
	fields := []any{
		"operation", operation,
		"error", err,
		"stack", string(debug.Stack()),
	}
	fields = append(fields, attrs...)
	slog.Error("p_seer_intel: aborted transaction trace", fields...)
}

func intelTableName(db *gorm.DB) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&Intel{}); err != nil {
		return "", err
	}
	return stmt.Schema.Table, nil
}
