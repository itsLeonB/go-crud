package gocrud_test

import (
	"context"
	"errors"
	"testing"

	"github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTransactorTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database")

	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "Failed to migrate test models")

	return db
}

func TestNewTransactor(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)

	assert.NotNil(t, transactor, "NewTransactor should not return nil")

	var _ ezutil.Transactor = transactor
}

func TestTransactor_Begin(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	ctx := context.Background()

	txCtx, err := transactor.Begin(ctx)
	assert.NoError(t, err, "Begin should not return error")
	assert.NotNil(t, txCtx, "Begin should not return nil context")

	// Verify transaction is in context
	tx, err := ezutil.GetTxFromContext(txCtx)
	assert.NoError(t, err, "GetTxFromContext should not return error")
	assert.NotNil(t, tx, "Begin should store transaction in context")
}

func TestTransactor_Commit(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	ctx := context.Background()

	t.Run("successful commit", func(t *testing.T) {
		txCtx, err := transactor.Begin(ctx)
		assert.NoError(t, err, "Failed to begin transaction")

		err = transactor.Commit(txCtx)
		assert.NoError(t, err, "Commit should not return error")
	})

	t.Run("commit without transaction", func(t *testing.T) {
		err := transactor.Commit(ctx)
		assert.NoError(t, err, "Commit should not error when no transaction exists")
	})
}

func TestTransactor_Rollback(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	ctx := context.Background()

	t.Run("successful rollback", func(t *testing.T) {
		txCtx, err := transactor.Begin(ctx)
		assert.NoError(t, err, "Failed to begin transaction")

		// Rollback should not panic or return error
		assert.NotPanics(t, func() {
			transactor.Rollback(txCtx)
		}, "Rollback should not panic")
	})

	t.Run("rollback without transaction", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			transactor.Rollback(ctx)
		}, "Rollback should not panic when no transaction exists")
	})
}

func TestTransactor_WithinTransaction_Success(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	repo := ezutil.NewCRUDRepository[TestModel](db)
	ctx := context.Background()

	var insertedID uint

	err := transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		model := TestModel{
			Name:  "Alice",
			Email: "alice@example.com",
			Age:   25,
		}

		result, err := repo.Insert(txCtx, model)
		if err != nil {
			return err
		}

		insertedID = result.ID
		return nil
	})

	assert.NoError(t, err, "WithinTransaction should not return error")

	// Verify the record was committed
	spec := ezutil.Specification[TestModel]{
		Model: TestModel{ID: insertedID},
	}
	result, err := repo.FindFirst(ctx, spec)
	assert.NoError(t, err, "Error verifying committed record")
	assert.NotZero(t, result.ID, "WithinTransaction record should be committed")
	assert.Equal(t, "Alice", result.Name, "Committed record should have correct data")
}

func TestTransactor_WithinTransaction_Rollback(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	repo := ezutil.NewCRUDRepository[TestModel](db)
	ctx := context.Background()

	var insertedID uint
	expectedError := errors.New("service error")

	err := transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		model := TestModel{
			Name:  "Bob",
			Email: "bob@example.com",
			Age:   30,
		}

		result, err := repo.Insert(txCtx, model)
		if err != nil {
			return err
		}

		insertedID = result.ID
		return expectedError
	})

	assert.Error(t, err, "WithinTransaction should return error")
	assert.Equal(t, expectedError, err, "WithinTransaction should return the expected error")

	// Verify the record was rolled back
	spec := ezutil.Specification[TestModel]{
		Model: TestModel{ID: insertedID},
	}
	result, err := repo.FindFirst(ctx, spec)
	assert.NoError(t, err, "Error checking rolled back record")
	assert.Zero(t, result.ID, "WithinTransaction record should be rolled back")
}

func TestTransactor_WithinTransaction_Nested(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	repo := ezutil.NewCRUDRepository[TestModel](db)
	ctx := context.Background()

	var outerID, innerID uint

	err := transactor.WithinTransaction(ctx, func(outerTxCtx context.Context) error {
		outerModel := TestModel{
			Name:  "Outer",
			Email: "outer@example.com",
			Age:   25,
		}

		result, err := repo.Insert(outerTxCtx, outerModel)
		if err != nil {
			return err
		}
		outerID = result.ID

		// Nested transaction (should reuse existing transaction)
		return transactor.WithinTransaction(outerTxCtx, func(innerTxCtx context.Context) error {
			innerModel := TestModel{
				Name:  "Inner",
				Email: "inner@example.com",
				Age:   30,
			}

			result, err := repo.Insert(innerTxCtx, innerModel)
			if err != nil {
				return err
			}
			innerID = result.ID

			return nil
		})
	})

	assert.NoError(t, err, "WithinTransaction nested should not return error")

	// Verify both records were committed
	outerSpec := ezutil.Specification[TestModel]{
		Model: TestModel{ID: outerID},
	}
	outerResult, err := repo.FindFirst(ctx, outerSpec)
	assert.NoError(t, err, "Error verifying outer record")
	assert.NotZero(t, outerResult.ID, "WithinTransaction outer record should be committed")

	innerSpec := ezutil.Specification[TestModel]{
		Model: TestModel{ID: innerID},
	}
	innerResult, err := repo.FindFirst(ctx, innerSpec)
	assert.NoError(t, err, "Error verifying inner record")
	assert.NotZero(t, innerResult.ID, "WithinTransaction inner record should be committed")
}

func TestGetTxFromContext(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := ezutil.NewTransactor(db)
	ctx := context.Background()

	t.Run("context with transaction", func(t *testing.T) {
		txCtx, err := transactor.Begin(ctx)
		assert.NoError(t, err, "Failed to begin transaction")

		tx, err := ezutil.GetTxFromContext(txCtx)
		assert.NoError(t, err, "GetTxFromContext should not return error")
		assert.NotNil(t, tx, "GetTxFromContext should return transaction")
	})

	t.Run("context without transaction", func(t *testing.T) {
		tx, err := ezutil.GetTxFromContext(ctx)
		assert.NoError(t, err, "GetTxFromContext should not return error")
		assert.Nil(t, tx, "GetTxFromContext should return nil when no transaction")
	})
}
