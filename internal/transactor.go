package internal

import (
	"context"
	"log"

	"github.com/itsLeonB/go-crud/lib"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type GormTransactor struct {
	DB *gorm.DB
}

func (t *GormTransactor) Begin(ctx context.Context) (context.Context, error) {
	tx := t.DB.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, eris.Wrap(err, lib.MsgTransactionError)
	}

	return context.WithValue(ctx, lib.ContextKeyGormTx, tx), nil
}

func (t *GormTransactor) Commit(ctx context.Context) error {
	tx, err := GetTxFromContext(ctx)
	if err != nil {
		return err
	}
	if tx != nil {
		err = tx.WithContext(ctx).Commit().Error
		if err != nil {
			return eris.Wrap(err, lib.MsgTransactionError)
		}
	}

	return nil
}

func (t *GormTransactor) Rollback(ctx context.Context) {
	tx, err := GetTxFromContext(ctx)
	if err != nil {
		log.Println("rollback error:", err)
		return
	}
	if tx == nil {
		log.Println("no transaction is running")
		return
	}

	err = tx.WithContext(ctx).Rollback().Error
	if err != nil {
		if err.Error() == "sql: transaction has already been committed or rolled back" {
			return
		}

		log.Printf("error: %T", err)
		log.Println("rollback error:", err)
	}
}

func (t *GormTransactor) WithinTransaction(ctx context.Context, serviceFn func(ctx context.Context) error) error {
	// Check if we're already within a transaction
	existingTx, err := GetTxFromContext(ctx)
	if err != nil {
		return eris.Wrap(err, "error checking existing transaction")
	}

	// If we're already in a transaction, just execute the service function
	if existingTx != nil {
		return serviceFn(ctx)
	}

	// Start a new transaction
	ctx, err = t.Begin(ctx)
	if err != nil {
		return eris.Wrap(err, "error starting transaction")
	}
	defer t.Rollback(ctx)

	if err := serviceFn(ctx); err != nil {
		return err
	}

	return t.Commit(ctx)
}

func GetTxFromContext(ctx context.Context) (*gorm.DB, error) {
	trx := ctx.Value(lib.ContextKeyGormTx)
	if trx != nil {
		tx, ok := trx.(*gorm.DB)
		if !ok {
			return nil, eris.New("error getting tx from ctx")
		}

		return tx, nil
	}

	return nil, nil
}
