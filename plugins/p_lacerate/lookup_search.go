package p_lacerate

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// formatFloat32VectorForPG returns a pgvector literal suitable for `?::vector` bindings.
func formatFloat32VectorForPG(vec []float32) string {
	if len(vec) == 0 {
		return intelZeroVectorTextForPG()
	}
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// searchTargetsOfInterestByEmbedding returns the nearest [TargetOfInterest] rows with non-null embeddings (cosine distance).
func searchTargetsOfInterestByEmbedding(db *gorm.DB, query []float32, limit int) ([]TargetOfInterest, error) {
	if len(query) != IntelEmbeddingDim {
		err := fmt.Errorf("query embedding dim %d, want %d", len(query), IntelEmbeddingDim)
		slog.Error("lacerate: search Targets of Interest by embedding", "error", err)
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	lit := formatFloat32VectorForPG(query)
	var rows []TargetOfInterest
	err := db.Raw(`
SELECT * FROM targets_of_interest
WHERE deleted_at IS NULL AND embedding IS NOT NULL
ORDER BY embedding <=> ?::vector ASC
LIMIT ?`, lit, limit).Scan(&rows).Error
	return rows, err
}

// searchIntelByEmbedding returns the nearest [Intel] rows by embedding (cosine distance).
func searchIntelByEmbedding(db *gorm.DB, query []float32, limit int) ([]Intel, error) {
	if len(query) != IntelEmbeddingDim {
		err := fmt.Errorf("query embedding dim %d, want %d", len(query), IntelEmbeddingDim)
		slog.Error("lacerate: search intel by embedding", "error", err)
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	lit := formatFloat32VectorForPG(query)
	var rows []Intel
	err := db.Raw(`
SELECT * FROM intels
WHERE deleted_at IS NULL
ORDER BY embedding <=> ?::vector ASC
LIMIT ?`, lit, limit).Scan(&rows).Error
	return rows, err
}
