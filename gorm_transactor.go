package ezutil

// todo move to go-crud pkg

import (
	"context"

	"github.com/itsLeonB/go-crud/internal"
	"gorm.io/gorm"
)

// Transactor provides an interface for managing database transactions with context.
// It abstracts transaction operations to allow for easier testing and different implementations.
type Transactor interface {
	// Begin starts a new database transaction and returns a context containing the transaction.
	Begin(ctx context.Context) (context.Context, error)
	// Commit commits the current transaction in the context.
	Commit(ctx context.Context) error
	// Rollback rolls back the current transaction in the context without returning an error.
	Rollback(ctx context.Context)
	// WithinTransaction executes a service function within a database transaction.
	WithinTransaction(ctx context.Context, serviceFn func(ctx context.Context) error) error
}

// NewTransactor creates a new Transactor implementation using GORM.
// The returned Transactor can be used to manage database transactions with context propagation.
func NewTransactor(db *gorm.DB) Transactor {
	return &internal.GormTransactor{DB: db}
}

// GetTxFromContext retrieves the current GORM transaction from the context.
// Returns an error if no transaction is found or if the stored value is not a *gorm.DB.
func GetTxFromContext(ctx context.Context) (*gorm.DB, error) {
	return internal.GetTxFromContext(ctx)
}
