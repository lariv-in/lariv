package p_totschool_tally

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// SessionsListGetter returns session id / display name pairs for the environment selector.
func SessionsListGetter(ctx context.Context) ([]registry.Pair[uint, string], error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
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
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return TotSchoolSession{}, err
	}
	session := getSessionFromEnvironment(db, ctx)
	return session, nil
}
