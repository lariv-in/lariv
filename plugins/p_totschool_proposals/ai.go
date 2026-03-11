package p_totschool_proposals

import (
	"context"
	"log"
	"time"

	"github.com/lariv-in/lago"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const (
	StatusPending    = "Pending"
	StatusGenerating = "Generating"
	StatusSuccess    = "Success"
	StatusFailed     = "Failed"
	StatusCanceled   = "Canceled"
)

// GenerationQueue is the in-process queue for AI generation (same schema as s_totschool_ai).
type GenerationQueue struct {
	ID                int64
	Content           string
	SystemPrompt      string         `gorm:"column:system_prompt"`
	Status            string         `gorm:"default:Pending"`
	GeneratedContent  *string        `gorm:"column:generated_content"`
	ErrorMessage      *string        `gorm:"column:error_message"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (GenerationQueue) TableName() string { return "generation_queue" }

var proposalDB *gorm.DB

func ensureGenerationQueue(d *gorm.DB) *gorm.DB {
	err := d.Exec(`
		CREATE TABLE IF NOT EXISTS generation_queue (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			content           TEXT NOT NULL,
			system_prompt     TEXT,
			status            TEXT NOT NULL DEFAULT 'Pending'
				CHECK (status IN ('Pending', 'Generating', 'Failed', 'Canceled', 'Success')),
			generated_content TEXT,
			error_message     TEXT,
			created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`).Error
	if err != nil {
		panic(err)
	}
	// Mark any stuck Generating as Failed on startup
	_ = d.Exec("UPDATE generation_queue SET status = 'Failed', error_message = 'Service restarted while generating' WHERE status = 'Generating'").Error
	return d
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		d = ensureGenerationQueue(d)
		proposalDB = d
		go runWorker()
		return d
	})
}

// EnqueueGeneration inserts a task and returns its ID (or 0 on error).
func EnqueueGeneration(content, systemPrompt string) (int64, error) {
	if proposalDB == nil {
		return 0, nil
	}
	task := GenerationQueue{
		Content:      content,
		SystemPrompt: systemPrompt,
		Status:       StatusPending,
	}
	err := proposalDB.Create(&task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

// GenerationStatus is the result of GetGenerationStatus.
type GenerationStatus struct {
	ID               int64
	Status           string
	GeneratedContent string
	ErrorMessage     string
}

// GetGenerationStatus returns status and optional content/error for a task.
func GetGenerationStatus(id int64) (GenerationStatus, bool) {
	if proposalDB == nil {
		return GenerationStatus{}, false
	}
	var t GenerationQueue
	err := proposalDB.Where("id = ?", id).First(&t).Error
	if err != nil {
		return GenerationStatus{}, false
	}
	out := GenerationStatus{ID: t.ID, Status: t.Status}
	if t.GeneratedContent != nil {
		out.GeneratedContent = *t.GeneratedContent
	}
	if t.ErrorMessage != nil {
		out.ErrorMessage = *t.ErrorMessage
	}
	return out, true
}

// CancelGeneration marks the task as Canceled.
func CancelGeneration(id int64) error {
	if proposalDB == nil {
		return nil
	}
	err := proposalDB.Model(&GenerationQueue{}).Where("id = ?", id).Updates(map[string]any{
		"status":      StatusCanceled,
		"updated_at": time.Now(),
	}).Error
	return err
}

func runWorker() {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{})
	if err != nil {
		log.Printf("[proposals] genai client not available (set GOOGLE_API_KEY or GEMINI_API_KEY): %v", err)
		return
	}

	for {
		time.Sleep(time.Second)
		if proposalDB == nil {
			continue
		}

		var task GenerationQueue
		err := proposalDB.Where("status = ?", StatusPending).First(&task).Error
		if err != nil || task.ID == 0 {
			continue
		}

		// Lock
		err = proposalDB.Model(&task).Update("status", StatusGenerating).Error
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		config := &genai.GenerateContentConfig{}
		if task.SystemPrompt != "" {
			config.SystemInstruction = genai.NewContentFromText(task.SystemPrompt, "user")
		}
		resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(task.Content), config)
		cancel()

		if err != nil {
			msg := err.Error()
			_ = proposalDB.Model(&task).Updates(map[string]any{
				"status":         StatusFailed,
				"error_message":  msg,
				"updated_at":     time.Now(),
			})
			continue
		}

		respText := resp.Text()
		if respText == "" {
			_ = proposalDB.Model(&task).Updates(map[string]any{
				"status":        StatusFailed,
				"error_message": "empty response",
				"updated_at":    time.Now(),
			})
			continue
		}

		// Check if was canceled while we were generating
		var current GenerationQueue
		_ = proposalDB.Where("id = ?", task.ID).First(&current)
		if current.Status == StatusCanceled {
			continue
		}

		_ = proposalDB.Model(&task).Updates(map[string]any{
			"status":             StatusSuccess,
			"generated_content": respText,
			"error_message":     nil,
			"updated_at":        time.Now(),
		})
	}
}
