package p_totschool_proposals

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
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
	Answers          datatypes.JSON // [{"question":"...","answer":"..."}, ...]
	GeneratedContent string         `gorm:"type:text"`
	GenerationID     *int           // non-nil while AI generation is in progress
}

// FormatAnswersForAI returns a single string of Q&A for the AI prompt.
func (p *Proposal) FormatAnswersForAI() (string, error) {
	if len(p.Answers) == 0 {
		return "", nil
	}

	// Answers are stored as []registry.Pair[string,string] via InputKeyValue.
	var items []registry.Pair[string, string]
	if err := json.Unmarshal(p.Answers, &items); err != nil {
		return "", err
	}

	var lines []string
	for i, item := range items {
		q := item.Key
		if q == "" && i < len(QUESTIONS) {
			q = QUESTIONS[i]
		}
		lines = append(lines, fmt.Sprintf("Q%d: %s\nA: %s\n", i+1, q, item.Value))
	}
	return strings.Join(lines, "\n"), nil
}

func init() {
	lago.OnDBInit("p_totschool_proposals.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Proposal](d)
		// Mark any stuck generating proposals as not generating on startup
		d.Model(&Proposal{}).Where("generation_id IS NOT NULL").Update("generation_id", nil)
		go runWorker(d)
		return d
	})
	lago.RegistryAdmin.Register("p_totschool_proposals", lago.AdminPanel[Proposal]{SearchField: "Title"})
}
