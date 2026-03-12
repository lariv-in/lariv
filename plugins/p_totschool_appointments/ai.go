package p_totschool_appointments

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	lago.RegistryConfig.Register("p_totschool_appointments", aiConfig)
}

type generationTask struct {
	AppointmentID uint
	Content       string
	SystemPrompt  string
}

var (
	workCh    = make(chan generationTask, 64)
	cancelMu  sync.Mutex
	cancelled = map[uint]bool{}
)

func Generate(db *gorm.DB, appointmentID uint, content, systemPrompt string) {
	one := 1
	db.Model(&Appointment{}).Where("id = ?", appointmentID).Updates(map[string]any{
		"generation_id":    &one,
		"generated_letter": "",
	})
	workCh <- generationTask{
		AppointmentID: appointmentID,
		Content:       content,
		SystemPrompt:  systemPrompt,
	}
}

func CancelGeneration(db *gorm.DB, appointmentID uint) {
	cancelMu.Lock()
	cancelled[appointmentID] = true
	cancelMu.Unlock()
	db.Model(&Appointment{}).Where("id = ?", appointmentID).Update("generation_id", nil)
}

func isCancelled(appointmentID uint) bool {
	cancelMu.Lock()
	defer cancelMu.Unlock()
	if cancelled[appointmentID] {
		delete(cancelled, appointmentID)
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
		log.Printf("[appointments] genai client not available: %v", err)
		for task := range workCh {
			db.Model(&Appointment{}).Where("id = ?", task.AppointmentID).Update("generation_id", nil)
		}
		return
	}

	for task := range workCh {
		if isCancelled(task.AppointmentID) {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		config := &genai.GenerateContentConfig{}
		if task.SystemPrompt != "" {
			config.SystemInstruction = genai.NewContentFromText(task.SystemPrompt, "user")
		}
		resp, err := client.Models.GenerateContent(ctx, model, genai.Text(task.Content), config)
		cancel()

		if isCancelled(task.AppointmentID) {
			continue
		}

		if err != nil {
			log.Printf("[appointments] generation failed for appointment %d: %v", task.AppointmentID, err)
			db.Model(&Appointment{}).Where("id = ?", task.AppointmentID).Update("generation_id", nil)
			continue
		}

		respText := resp.Text()
		if respText == "" {
			log.Printf("[appointments] empty response for appointment %d", task.AppointmentID)
			db.Model(&Appointment{}).Where("id = ?", task.AppointmentID).Update("generation_id", nil)
			continue
		}

		db.Model(&Appointment{}).Where("id = ?", task.AppointmentID).Updates(map[string]any{
			"generated_letter": respText,
			"generation_id":    nil,
		})
	}
}

const letterWriterSystemPrompt = `You are an expert letter writer. Your task is to rewrite or generate the given letter while incorporating any provided remarks/notes.

Rules:
1. Only output the letter content - no explanations, no markdown formatting, no code blocks
2. Preserve all factual information (names, dates, times, locations) from the original letter
3. Incorporate any remarks naturally into the letter's tone, style, or content as appropriate
4. Maintain a professional tone unless the remarks suggest otherwise
5. Output plain text only - no special formatting
6. If no suitable output can be produced, write a sample template`

const letterEditorSystemPrompt = `You are an expert letter editor. Your task is to edit or rewrite the given letter according to the user's instructions.

Rules:
1. Only output the edited letter content - no explanations, no markdown formatting, no code blocks
2. Preserve the general structure and intent of the letter unless instructed otherwise
3. Maintain a professional tone unless instructed otherwise
4. Keep all factual information (names, dates, times, locations) unchanged unless specifically asked to modify them
5. Output plain text only - no special formatting
6. If no suitable output can be produced, write a sample template`

func buildLetterContent(db *gorm.DB, appointment *Appointment, userName string) string {
	var template LetterTemplate
	if err := db.First(&template).Error; err != nil {
		return fmt.Sprintf("Write a professional appointment letter for %s.", appointment.Name)
	}

	content := template.Content
	replacements := map[string]string{
		"{{appointment_name}}":            appointment.Name,
		"{{appointment_location}}":        appointment.Location,
		"{{appointment_datetime}}":        appointment.Datetime.Format("January 02, 2006 03:04 PM"),
		"{{appointment_phone}}":           appointment.Phone,
		"{{appointment_extra_info}}":      appointment.ExtraInfo,
		"{{appointment_created_at}}":      appointment.CreatedAt.Format("January 02, 2006 03:04 PM"),
		"{{appointment_created_at_date}}": appointment.CreatedAt.Format("January 02, 2006"),
		"{{user_name}}":                   userName,
	}

	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return content
}
