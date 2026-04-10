package p_lacerate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubVLEmbedder struct {
	vec []float32
	err error
}

func (s stubVLEmbedder) Embed(ctx context.Context, text string, images ...[]byte) ([]float32, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]float32, len(s.vec))
	copy(out, s.vec)
	return out, nil
}

func lacerateTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	if err := db.AutoMigrate(&p_filesystem.VNode{}, &Source{}, &Intel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	return db
}

const lacerateTestPostgresDSNEnv = "LACERATE_TEST_POSTGRES_DSN"

func lacerateTestPostgresDB(t *testing.T) *gorm.DB {
	t.Helper()

	baseDSN := strings.TrimSpace(os.Getenv(lacerateTestPostgresDSNEnv))
	if baseDSN == "" {
		t.Skipf("%s is not set", lacerateTestPostgresDSNEnv)
	}

	adminDB, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{PrepareStmt: true})
	if err != nil {
		t.Fatalf("gorm.Open postgres failed: %v", err)
	}

	schema := fmt.Sprintf("lacerate_test_%d", time.Now().UnixNano())
	if err := adminDB.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)).Error; err != nil {
		t.Fatalf("create schema failed: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE`, schema)).Error
	})

	if err := adminDB.Exec(`CREATE EXTENSION IF NOT EXISTS vector`).Error; err != nil {
		t.Fatalf("create pgvector extension failed: %v", err)
	}

	testDSN := baseDSN
	if strings.Contains(testDSN, "?") {
		testDSN += "&search_path=" + schema + ",public"
	} else {
		testDSN += "?search_path=" + schema + ",public"
	}
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{PrepareStmt: true})
	if err != nil {
		t.Fatalf("gorm.Open postgres test db failed: %v", err)
	}
	if err := db.AutoMigrate(&p_filesystem.VNode{}, &Source{}, &Intel{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	return db
}

func testEmbeddingVector() []float32 {
	vec := make([]float32, IntelEmbeddingDim)
	vec[0] = 1.25
	vec[17] = -0.5
	vec[101] = 0.75
	return vec
}

func TestIntelCreatePersistsComputedEmbedding(t *testing.T) {
	db := lacerateTestDB(t)
	testIntelCreatePersistsComputedEmbedding(t, db)
}

func TestIntelCreatePersistsComputedEmbeddingPostgres(t *testing.T) {
	db := lacerateTestPostgresDB(t)
	testIntelCreatePersistsComputedEmbedding(t, db)
}

func testIntelCreatePersistsComputedEmbedding(t *testing.T, db *gorm.DB) {
	previous := vlEmbedder()
	RegisterVLEmbedder(stubVLEmbedder{vec: testEmbeddingVector()})
	t.Cleanup(func() {
		RegisterVLEmbedder(previous)
	})

	src := Source{Name: "test", Kind: "reddit"}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("Create source failed: %v", err)
	}

	dedup := "dedup"
	intel := Intel{
		SourceID:  src.ID,
		DedupHash: &dedup,
		Content:   "hello world",
	}
	if err := db.Create(&intel).Error; err != nil {
		t.Fatalf("Create intel failed: %v", err)
	}

	var got Intel
	if err := db.First(&got, intel.ID).Error; err != nil {
		t.Fatalf("Load intel failed: %v", err)
	}

	vec := got.Embedding.Slice()
	if len(vec) != IntelEmbeddingDim {
		t.Fatalf("expected embedding dim %d, got %d", IntelEmbeddingDim, len(vec))
	}
	if vec[0] != 1.25 || vec[17] != -0.5 || vec[101] != 0.75 {
		t.Fatalf("expected persisted embedding values, got [%v %v %v]", vec[0], vec[17], vec[101])
	}
	if nonZero, _ := embeddingStats(vec); nonZero != 3 {
		t.Fatalf("expected 3 non-zero dimensions, got %d", nonZero)
	}
}
