package views

import "gorm.io/gorm"

// TxCommitHook defines an interface for GORM model types that execute hooks after a creation or update transaction successfully commits.
// Unlike GORM's built-in transaction hooks (which run within the active transaction blocks),
// the hook parameter represents the request-scoped pooled db connection (running outside the transaction context).
//
// Use Cases:
//   - Sending confirmation/welcome emails immediately after a user record has committed to the database.
//   - Triggering async operations (like publishing event payloads, starting workers, or invalidating caches) post-commit.
//
// Example:
//
//	type Member struct {
//		gorm.Model
//		Email string
//	}
//
//	func (m *Member) AfterTxCommit(db *gorm.DB) {
//		// Safely queue welcome notification after transaction commits.
//		go queueWelcomeEmail(m.Email)
//	}
type TxCommitHook interface {
	// AfterTxCommit runs logic immediately after a successful transaction commit.
	AfterTxCommit(db *gorm.DB)
}
