package p_totschool_tally

import (
	"context"
	"time"

	"github.com/lariv-in/lago/getters"
)

// tallySessionEnvironmentDefault returns the TotSchoolSession id for the current quarter.
func tallySessionEnvironmentDefault(ctx context.Context) (uint, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return 0, err
	}
	session := EnsureSessionForDate(db, time.Now())
	return session.ID, nil
}
