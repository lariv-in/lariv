package p_totschool_proposals

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/lariv-in/p_users"
	"github.com/lariv-in/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var proposalTitles = []string{
	"Retirement Planning Proposal",
	"Child Education Plan",
	"Comprehensive Financial Plan",
	"Wealth Management Proposal",
	"Insurance Portfolio Review",
	"Tax Optimization Strategy",
	"Investment Strategy",
	"Estate Planning Outline",
	"Debt Reduction Plan",
	"Savings and Investment Proposal",
	"Family Financial Review",
	"Risk Management Assessment",
	"Future Aspirations Plan",
}

// GenerateProposalsForUser creates count proposals for the given user with random answers.
func GenerateProposalsForUser(db *gorm.DB, user *p_users.User, count int) (int, error) {
	created := 0
	for created < count {
		// Build answers as []registry.Pair[string,string] to match InputKeyValue
		// and FieldKeyValue JSON schema.
		answers := make([]registry.Pair[string, string], len(QUESTIONS))
		for i, q := range QUESTIONS {
			answers[i] = registry.Pair[string, string]{
				Key:   q,
				Value: randomSentence(5, 15),
			}
		}
		titleIdx := rand.Intn(len(proposalTitles))
		title := fmt.Sprintf("%s - User %d", proposalTitles[titleIdx], user.ID)

		b, err := json.Marshal(answers)
		if err != nil {
			return created, err
		}
		p := Proposal{
			CreatedByID: user.ID,
			Title:       title,
			Answers:     datatypes.JSON(b),
		}
		if err := db.Create(&p).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func randomSentence(minWords, maxWords int) string {
	words := []string{"The", "client", "has", "stated", "that", "they", "want", "to", "save", "for", "future", "and", "invest", "in", "property", "with", "adequate", "insurance", "coverage", "for", "family."}
	n := minWords + rand.Intn(maxWords-minWords+1)
	if n > len(words) {
		n = len(words)
	}
	s := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			s += " "
		}
		s += words[rand.Intn(len(words))]
	}
	return s + "."
}
