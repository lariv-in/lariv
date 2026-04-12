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
SELECT * FROM reports
WHERE deleted_at IS NULL AND embedding IS NOT NULL
ORDER BY embedding <=> ? ASC
LIMIT ?`, query, limit).Scan(&rows).Error
	return rows, err
}

// searchTargetsOfInterestByEmbedding returns the nearest [TargetOfInterest] rows with non-null embeddings (cosine distance).
func searchTargetsOfInterestByEmbedding(db *gorm.DB, query pgvector.Vector, limit int) ([]TargetOfInterest, error) {
	qslice := query.Slice()
	if len(qslice) != IntelEmbeddingDim {
		err := fmt.Errorf("query embedding dim %d, want %d", len(qslice), IntelEmbeddingDim)
		slog.Error("lacerate: search targets of interest by embedding", "error", err)
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []TargetOfInterest
	err := db.Raw(`
SELECT * FROM targets_of_interest
WHERE deleted_at IS NULL AND embedding IS NOT NULL
ORDER BY embedding <=> ? ASC
LIMIT ?`, query, limit).Scan(&rows).Error
	return rows, err
}

// nearestTargetOfInterestByEmbedding returns the single nearest [TargetOfInterest] by cosine distance
// (pgvector <=>: 0 = identical, 2 = opposite). ok is false when no row has a non-null embedding.
func nearestTargetOfInterestByEmbedding(db *gorm.DB, query pgvector.Vector) (t TargetOfInterest, cosineDistance float64, ok bool, err error) {
	qslice := query.Slice()
	if len(qslice) != IntelEmbeddingDim {
		err = fmt.Errorf("query embedding dim %d, want %d", len(qslice), IntelEmbeddingDim)
		slog.Error("lacerate: nearest target of interest by embedding", "error", err)
		return TargetOfInterest{}, 0, false, err
	}
	var row struct {
		TargetOfInterest `gorm:"embedded"`
		CosineDistance   float64 `gorm:"column:cosine_dist"`
	}
	err = db.Raw(`
SELECT t.*, (t.embedding <=> ?)::float8 AS cosine_dist
FROM targets_of_interest t
WHERE t.deleted_at IS NULL AND t.embedding IS NOT NULL
ORDER BY t.embedding <=> ? ASC
LIMIT 1`, query, query).Scan(&row).Error
	if err != nil {
		slog.Error("lacerate: nearest target of interest by embedding", "error", err)
		return TargetOfInterest{}, 0, false, err
	}
	if row.TargetOfInterest.ID == 0 {
		return TargetOfInterest{}, 0, false, nil
	}
	return row.TargetOfInterest, row.CosineDistance, true, nil
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
