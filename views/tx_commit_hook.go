package views

import "gorm.io/gorm"

// TxCommitHook is implemented by model types that need work after a successful
// LayerCreate or LayerUpdate transaction. db is the request-scoped pooled *gorm.DB
// (not the transactional *gorm.DB passed to GORM hooks).
type TxCommitHook interface {
	AfterTxCommit(db *gorm.DB)
}
