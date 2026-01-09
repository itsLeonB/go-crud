package crud_test

import (
	"context"
	"errors"
	"testing"

	crud "github.com/itsLeonB/go-crud"
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

	err = db.AutoMigrate(&TestUser{})
	assert.NoError(t, err, "Failed to migrate test models")

	return db
}

func TestNewTransactor(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)

	assert.NotNil(t, transactor, "NewTransactor should not return nil")
}

func TestTransactor_Begin(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)
	ctx := context.Background()

	txCtx, err := transactor.Begin(ctx)
	assert.NoError(t, err, "Begin should not return error")
	assert.NotNil(t, txCtx, "Begin should not return nil context")
}

func TestTransactor_Commit(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)
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
	transactor := crud.NewTransactor(db)
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
	transactor := crud.NewTransactor(db)
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	var insertedID uint

	err := transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		model := TestUser{
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
	spec := crud.Specification[TestUser]{
		Model: TestUser{ID: insertedID},
	}
	result, err := repo.FindFirst(ctx, spec)
	assert.NoError(t, err, "Error verifying committed record")
	assert.NotZero(t, result.ID, "WithinTransaction record should be committed")
	assert.Equal(t, "Alice", result.Name, "Committed record should have correct data")
}

func TestTransactor_WithinTransaction_Rollback(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	expectedError := errors.New("service error")

	err := transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		model := TestUser{
			Name:  "Bob",
			Email: "bob@example.com",
			Age:   30,
		}

		_, err := repo.Insert(txCtx, model)
		if err != nil {
			return err
		}

		return expectedError
	})

	assert.Error(t, err, "WithinTransaction should return error")
	assert.Equal(t, expectedError, err, "WithinTransaction should return the expected error")

	// Verify no records exist (rollback worked)
	spec := crud.Specification[TestUser]{}
	results, err := repo.FindAll(ctx, spec)
	assert.NoError(t, err)
	assert.Len(t, results, 0, "WithinTransaction should rollback all records")
}

func TestTransactor_WithinTransaction_Nested(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	var outerID, innerID uint

	err := transactor.WithinTransaction(ctx, func(outerTxCtx context.Context) error {
		outerModel := TestUser{
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
			innerModel := TestUser{
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
	outerSpec := crud.Specification[TestUser]{
		Model: TestUser{ID: outerID},
	}
	outerResult, err := repo.FindFirst(ctx, outerSpec)
	assert.NoError(t, err, "Error verifying outer record")
	assert.NotZero(t, outerResult.ID, "WithinTransaction outer record should be committed")

	innerSpec := crud.Specification[TestUser]{
		Model: TestUser{ID: innerID},
	}
	innerResult, err := repo.FindFirst(ctx, innerSpec)
	assert.NoError(t, err, "Error verifying inner record")
	assert.NotZero(t, innerResult.ID, "WithinTransaction inner record should be committed")
}

func TestTransactor_Begin_Error(t *testing.T) {
	// Create a closed database to trigger error
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	err = sqlDB.Close()
	assert.NoError(t, err)

	transactor := crud.NewTransactor(db)
	ctx := context.Background()

	_, err = transactor.Begin(ctx)
	assert.Error(t, err)
}

func TestTransactor_Commit_Error(t *testing.T) {
	db := setupTransactorTestDB(t)
	transactor := crud.NewTransactor(db)
	ctx := context.Background()

	// Test committing without a transaction (should not error)
	err := transactor.Commit(ctx)
	assert.NoError(t, err) // This is expected behavior
}
