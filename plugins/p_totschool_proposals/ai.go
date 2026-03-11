package p_totschool_proposals

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lariv-in/lago"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

type AIConfig struct {
	APIKey string `toml:"apiKey"`
	Model  string `toml:"model"`
}

var aiConfig = &AIConfig{}

func (c *AIConfig) PostConfig() {}

func init() {
	lago.RegistryConfig.Register("p_totschool_proposals", aiConfig)
}

// generationTask is sent through the work channel to the AI worker.
type generationTask struct {
	ProposalID   uint
	Content      string
	SystemPrompt string
}

var (
	workCh    = make(chan generationTask, 64)
	cancelMu  sync.Mutex
	cancelled = map[uint]bool{}
)

// Generate sends a generation task to the worker goroutine.
func Generate(db *gorm.DB, proposalID uint, content, systemPrompt string) {
	// Mark proposal as generating
	one := 1
	db.Model(&Proposal{}).Where("id = ?", proposalID).Updates(map[string]any{
		"generation_id":     &one,
		"generated_content": "",
	})
	workCh <- generationTask{
		ProposalID:   proposalID,
		Content:      content,
		SystemPrompt: systemPrompt,
	}
}

// CancelGeneration marks a proposal's generation as cancelled.
func CancelGeneration(db *gorm.DB, proposalID uint) {
	cancelMu.Lock()
	cancelled[proposalID] = true
	cancelMu.Unlock()
	db.Model(&Proposal{}).Where("id = ?", proposalID).Update("generation_id", nil)
}

func isCancelled(proposalID uint) bool {
	cancelMu.Lock()
	defer cancelMu.Unlock()
	if cancelled[proposalID] {
		delete(cancelled, proposalID)
		return true
	}
	return false
}

func runWorker(db *gorm.DB) {
	clientConfig := &genai.ClientConfig{}
	if aiConfig.APIKey != "" {
		clientConfig.APIKey = aiConfig.APIKey
	}
	model := "gemini-2.5-flash"
	if aiConfig.Model != "" {
		model = aiConfig.Model
	}

	client, err := genai.NewClient(context.Background(), clientConfig)
	if err != nil {
		log.Printf("[proposals] genai client not available: %v", err)
		// Drain channel so senders don't block, mark proposals as failed
		for task := range workCh {
			db.Model(&Proposal{}).Where("id = ?", task.ProposalID).Update("generation_id", nil)
		}
		return
	}

	for task := range workCh {
		if isCancelled(task.ProposalID) {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		config := &genai.GenerateContentConfig{}
		if task.SystemPrompt != "" {
			config.SystemInstruction = genai.NewContentFromText(task.SystemPrompt, "user")
		}
		resp, err := client.Models.GenerateContent(ctx, model, genai.Text(task.Content), config)
		cancel()

		if isCancelled(task.ProposalID) {
			continue
		}

		if err != nil {
			log.Printf("[proposals] generation failed for proposal %d: %v", task.ProposalID, err)
			db.Model(&Proposal{}).Where("id = ?", task.ProposalID).Update("generation_id", nil)
			continue
		}

		respText := resp.Text()
		if respText == "" {
			log.Printf("[proposals] empty response for proposal %d", task.ProposalID)
			db.Model(&Proposal{}).Where("id = ?", task.ProposalID).Update("generation_id", nil)
			continue
		}

		db.Model(&Proposal{}).Where("id = ?", task.ProposalID).Updates(map[string]any{
			"generated_content": respText,
			"generation_id":     nil,
		})
	}
}
