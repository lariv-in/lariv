package p_totschool_tally

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// tallySessionEnvironmentDefault returns the TotSchoolSession id for the current quarter.
func tallySessionEnvironmentDefault(ctx context.Context) (uint, error) {
	db := ctx.Value("$db").(*gorm.DB)
	session := EnsureSessionForDate(db, time.Now())
	return session.ID, nil
}

// SessionsListGetter returns session id / display name pairs for the environment selector.
func SessionsListGetter(ctx context.Context) ([]registry.Pair[uint, string], error) {
	db := ctx.Value("$db").(*gorm.DB)
	sessions, err := gorm.G[TotSchoolSession](db).Order(`"start" DESC`).Find(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]registry.Pair[uint, string], 0, len(sessions))
	for _, s := range sessions {
		out = append(out, registry.Pair[uint, string]{
			Key:   s.ID,
			Value: s.Name,
		})
	}
	return out, nil
}

// CurrentEnvironmentSessionGetter resolves the active TotSchoolSession from
// the $environment cookie (or falls back to the current quarter), matching
// the behaviour used on tally dashboard/list pages.
func CurrentEnvironmentSessionGetter(ctx context.Context) (TotSchoolSession, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return TotSchoolSession{}, fmt.Errorf("TallySessionEntries: missing $db in context")
	}
	session := getSessionFromEnvironment(db, ctx)
	return session, nil
}

func tallyCommonFields() []components.PageInterface {
	return []components.PageInterface{
		components.ContainerRow{
			Classes: "grid grid-cols-1 md:grid-cols-2 gap-4",
			Children: []components.PageInterface{
				components.InputNumber[int]{Name: "Visits", Label: "Visits", Required: true, Getter: getters.Key[int]("$in.Visits")},
				components.InputNumber[int]{Name: "Appointments", Label: "Appointments", Required: true, Getter: getters.Key[int]("$in.Appointments")},
				components.InputNumber[int]{Name: "Leads", Label: "Leads", Required: true, Getter: getters.Key[int]("$in.Leads")},
				components.InputNumber[int]{Name: "Presentations", Label: "Presentations", Required: true, Getter: getters.Key[int]("$in.Presentations")},
				components.InputNumber[int]{Name: "Demos", Label: "Demonstrations", Required: true, Getter: getters.Key[int]("$in.Demos")},
				components.InputNumber[int]{Name: "Letters", Label: "Follow Up Letters Sent", Required: true, Getter: getters.Key[int]("$in.Letters")},
				components.InputNumber[int]{Name: "FollowUps", Label: "Follow Ups", Required: true, Getter: getters.Key[int]("$in.FollowUps")},
				components.InputNumber[int]{Name: "Proposals", Label: "Proposals Given", Required: true, Getter: getters.Key[int]("$in.Proposals")},
				components.InputNumber[int]{Name: "Policies", Label: "Policies Sold", Required: true, Getter: getters.Key[int]("$in.Policies")},
				components.InputNumber[int]{Name: "Premium", Label: "Premium", Required: true, Getter: getters.Key[int]("$in.Premium")},
			},
		},
	}
}
