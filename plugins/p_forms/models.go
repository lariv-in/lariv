package forms

import (
	"encoding/json"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FieldTypeChoices is stored field type Key -> admin label Value (slice order = dropdown order).
var FieldTypeChoices = []registry.Pair[string, string]{
	{Key: "email", Value: "Email"},
	{Key: "number", Value: "Number"},
	{Key: "select", Value: "Select"},
	{Key: "text", Value: "Text"},
	{Key: "textarea", Value: "Textarea"},
}

// Form is a form definition (title, URL slug, optional description).
type Form struct {
	gorm.Model
	Title       string `gorm:"size:250;not null"`
	Slug        string `gorm:"size:160;uniqueIndex;not null"`
	Description string `gorm:"type:text"`

	FormFields []FormField `gorm:"constraint:OnDelete:CASCADE;"`
}

// FormField is one input in a form (visual builder row).
type FormField struct {
	gorm.Model
	FormID    uint   `gorm:"not null;uniqueIndex:idx_form_field_unique,priority:1"`
	Form      Form   `gorm:"constraint:OnDelete:CASCADE;"`
	SortOrder int    `gorm:"default:0"`
	Name      string `gorm:"size:120;not null;uniqueIndex:idx_form_field_unique,priority:2"`
	Label     string `gorm:"size:250;not null"`
	FieldType string `gorm:"size:32;not null"`
	Required  bool
	// Options is a JSON-encoded []string of choice values when FieldType is "select".
	Options string `gorm:"type:text"`
}

// FormSubmission stores one public submit payload as JSON (object: field name -> value).
type FormSubmission struct {
	gorm.Model
	FormID  uint           `gorm:"not null;index"`
	Form    Form           `gorm:"constraint:OnDelete:CASCADE;"`
	Answers datatypes.JSON `gorm:"type:jsonb;not null"`
}

// SelectOptionStrings returns select choices from Options as a JSON array of strings.
func (f *FormField) SelectOptionStrings() []string {
	s := strings.TrimSpace(f.Options)
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil
	}
	return arr
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Form](d)
		lago.RegisterModel[FormField](d)
		lago.RegisterModel[FormSubmission](d)
		return d
	})

	lago.RegistryAdmin.Register("forms", lago.AdminPanel[Form]{
		SearchField: "Title",
		ListFields: []string{
			"Title",
			"Slug",
			"Description",
			"CreatedAt",
			"UpdatedAt",
		},
	})
}
