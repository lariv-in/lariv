package getters

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DBFromContext returns the GORM DB injected by app HTTP layers (context value ContextKeyDB).
func DBFromContext(ctx context.Context) (*gorm.DB, error) {
	db, ok := ctx.Value(ContextKeyDB).(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("getters.DBFromContext: missing or nil %s", ContextKeyDB)
	}
	return db, nil
}
