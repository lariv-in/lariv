package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// LookupTargetOfInterestTouchActionChoices labels tool actions stored on [LookupLogTargetOfInterest.Action].
var LookupTargetOfInterestTouchActionChoices = []registry.Pair[string, string]{
	{Key: "create", Value: "Created"},
	{Key: "edit", Value: "Edited"},
}

// LookupLogTargetOfInterest records a successful create/edit of a [TargetOfInterest] from a lookup agent tool call.
type LookupLogTargetOfInterest struct {
	gorm.Model
	LookupID uint   `gorm:"not null;index"`
	Lookup   Lookup `gorm:"foreignKey:LookupID;constraint:OnDelete:CASCADE"`
	// LookupLogEntryID is the sibling [LookupLogEntry] row (Kind tool_call) for this touch.
	LookupLogEntryID   uint             `gorm:"not null;index"`
	LookupLogEntry     LookupLogEntry   `gorm:"foreignKey:LookupLogEntryID;constraint:OnDelete:CASCADE"`
	TargetOfInterestID uint             `gorm:"not null;index"`
	TargetOfInterest   TargetOfInterest `gorm:"foreignKey:TargetOfInterestID"`
	// Action is "create" or "edit" (keys in [LookupTargetOfInterestTouchActionChoices]).
	Action string `gorm:"size:16;not null"`
}

func (LookupLogTargetOfInterest) TableName() string { return "lookup_log_targets_of_interest" }

func init() {
	lago.OnDBInit("p_lacerate.lookup_log_target_of_interest", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[LookupLogTargetOfInterest](db)
		return db
	})
}
