package p_lacerate

import (
	"fmt"
	"log/slog"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// searchReportsByEmbedding returns the nearest [Report] rows with non-null embeddings (cosine distance).
func searchReportsByEmbedding(db *gorm.DB, query pgvector.Vector, limit int) ([]Report, error) {
	qslice := query.Slice()
	if len(qslice) != IntelEmbeddingDim {
		err := fmt.Errorf("query embedding dim %d, want %d", len(qslice), IntelEmbeddingDim)
		slog.Error("lacerate: search reports by embedding", "error", err)
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []Report
	err := db.Raw(`
SELECT * FROM targets_of_interest
WHERE deleted_at IS NULL AND embedding IS NOT NULL
ORDER BY embedding <=> ? ASC
LIMIT ?`, query, limit).Scan(&rows).Error
	return rows, err
}

// searchIntelByEmbedding returns the nearest [Intel] rows by embedding (cosine distance).
func searchIntelByEmbedding(db *gorm.DB, query pgvector.Vector, limit int) ([]Intel, error) {
	qslice := query.Slice()
	if len(qslice) != IntelEmbeddingDim {
		err := fmt.Errorf("query embedding dim %d, want %d", len(qslice), IntelEmbeddingDim)
		slog.Error("lacerate: search intel by embedding", "error", err)
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []Intel
	err := db.Raw(`
SELECT * FROM intels
WHERE deleted_at IS NULL
ORDER BY embedding <=> ? ASC
LIMIT ?`, query, limit).Scan(&rows).Error
	return rows, err
}
