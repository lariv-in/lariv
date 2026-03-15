package p_totschool_appointments

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
	"gorm.io/gorm"
)

type GenerationTask struct {
	AppointmentID uint
	Content       string
	SystemPrompt  string
}

var (
	workCh    = make(chan GenerationTask, 64)
	cancelMu  sync.Mutex
	cancelled = map[uint]bool{}
)

func Generate(db *gorm.DB, appointmentID uint, content, systemPrompt string) {
	one := 1
	db.Model(&Appointment{}).Where("id = ?", appointmentID).Updates(map[string]any{
		"generation_id":    &one,
		"generated_letter": "",
	})
	workCh <- GenerationTask{
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
	if totschoolAppointmentConfig.APIKey != "" {
		clientConfig.APIKey = totschoolAppointmentConfig.APIKey
	}
	model := "gemini-2.5-flash"
	if totschoolAppointmentConfig.Model != "" {
		model = totschoolAppointmentConfig.Model
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

const letterWriterSystemPrompt = `You are an expert letter writer. You will receive:
1. A letter draft with some placeholders already filled (appointment name, dates, user name).
2. Separately: the meeting location and any extra info/context.

Your task: Produce the final letter by incorporating the location and extra info naturally into the draft. Weave in where the meeting will take place and any extra context in an appropriate place—do not just paste them as separate lines.

Rules:
1. Only output the final letter content - no explanations, no markdown formatting, no code blocks
2. Keep all factual information from the draft (names, dates, times) as given
3. Incorporate the provided location naturally (e.g. "We've blocked at ... at [location] for our meeting")
4. Incorporate the extra info naturally into the letter's tone or content where it fits
5. Maintain a professional tone
6. Output plain text only - no special formatting`

const letterEditorSystemPrompt = `You are an expert letter editor. Your task is to edit or rewrite the given letter according to the user's instructions.

Rules:
1. Only output the edited letter content - no explanations, no markdown formatting, no code blocks
2. Preserve the general structure and intent of the letter unless instructed otherwise
3. Maintain a professional tone unless instructed otherwise
4. Keep all factual information (names, dates, times, locations) unchanged unless specifically asked to modify them
5. Output plain text only - no special formatting
6. If no suitable output can be produced, write a sample template`

// Single letter template. Only direct factual placeholders are filled here;
// location and extra_info are passed to the AI separately for natural incorporation.
const letterTemplate = `Dear {{appointment_name}} SIR,

Thank you for the brief conversation on {{appointment_created_at_date}}. It was a pleasure interacting with you to exchanging my brochure.
We've blocked at {{appointment_datetime}} for our next meeting.

I'm looking forward to showing you the wealth-building demonstrations and financial wisdom concepts we discussed.


I am attaching my brochure below so you can get a reminder of our meeting for the work.

See you as per the decided schedule and thank you once again


{{user_name}}
Family Wealth Educator
`

func buildLetterContent(db *gorm.DB, appointment *Appointment, userName string) (userContent, systemPrompt string) {
	replacements := map[string]string{
		"{{appointment_name}}":            appointment.Name,
		"{{appointment_datetime}}":        appointment.Datetime.Format("January 02, 2006 03:04 PM"),
		"{{appointment_created_at_date}}": appointment.CreatedAt.Format("January 02, 2006"),
		"{{user_name}}":                   userName,
	}
	content := letterTemplate
	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// Pass location and extra_info to the AI separately so it can incorporate them naturally.
	userContent = "Letter draft:\n\n" + content
	if appointment.Location != "" {
		userContent += "\n\nMeeting location (incorporate naturally into the letter): " + appointment.Location
	}
	if appointment.ExtraInfo != "" {
		userContent += "\n\nExtra context to incorporate naturally: " + appointment.ExtraInfo
	}
	return userContent, letterWriterSystemPrompt
}
