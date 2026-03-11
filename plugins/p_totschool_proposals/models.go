package p_totschool_proposals

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var QUESTIONS = []string{
	"How many members are there in your family?",
	"What are their names and ages?",
	"What is your total monthly income?",
	"What is your total monthly expenditure?",
	"Where do you save money?",
	"If you stay out of station for a whole month, what would be the family's expenditure?",
	"What kind of insurance policies do you have?",
	"Do you have any additional source of income?",
	"Do you have any property?",
	"Are there any loans pending on your properties?",
	"Do you have any other liabilities?",
	"What shall be your income 5 years later?",
	"What shall be your expenditure 5 years later?",
	"What shall be your income 10 years later?",
	"What shall be your expenditure 10 years later?",
	"What are your plans for your children's future aspirations?",
	"What are your plans for your retired life?",
	"What is your dream car vs present vehicle?",
	"What is your dream international vacation?",
	"Give me details of your dream home.",
	"What is your dream monthly income.",
	"What is your dream monthly pension.",
}

type Proposal struct {
	gorm.Model
	CreatedByID      uint           `gorm:"notnull"`
	CreatedBy        p_users.User   `gorm:"foreignKey:CreatedByID"`
	Title            string         `gorm:"size:250;notnull"`
	Answers          datatypes.JSON `gorm:"type:text"` // [{"question":"...","answer":"..."}, ...]
	GeneratedContent string         `gorm:"type:text"`
	GenerationID     *int           // non-nil while AI generation is in progress
}

// QAItem is one question-answer pair for JSON (Answers).
type QAItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ParseAnswers deserializes Answers JSON into []QAItem.
func (p *Proposal) ParseAnswers() ([]QAItem, error) {
	if len(p.Answers) == 0 {
		return nil, nil
	}
	var out []QAItem
	err := json.Unmarshal(p.Answers, &out)
	return out, err
}

// SetAnswers serializes []QAItem into Answers.
func (p *Proposal) SetAnswers(items []QAItem) error {
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	p.Answers = datatypes.JSON(b)
	return nil
}

// FormatAnswersForAI returns a single string of Q&A for the AI prompt.
func (p *Proposal) FormatAnswersForAI() (string, error) {
	items, err := p.ParseAnswers()
	if err != nil {
		return "", err
	}
	var lines []string
	for i, item := range items {
		q := item.Question
		if q == "" && i < len(QUESTIONS) {
			q = QUESTIONS[i]
		}
		lines = append(lines, fmt.Sprintf("Q%d: %s\nA: %s\n", i+1, q, item.Answer))
	}
	return strings.Join(lines, "\n"), nil
}

// InitializeAnswers sets Answers to QUESTIONS with empty answers.
func (p *Proposal) InitializeAnswers() {
	items := make([]QAItem, len(QUESTIONS))
	for i, q := range QUESTIONS {
		items[i] = QAItem{Question: q, Answer: ""}
	}
	_ = p.SetAnswers(items)
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Proposal{}); err != nil {
			panic(err)
		}
		// Mark any stuck generating proposals as not generating on startup
		d.Model(&Proposal{}).Where("generation_id IS NOT NULL").Update("generation_id", nil)
		go runWorker(d)
		return d
	})
}
