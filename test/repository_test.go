package gocrud_test

import (
	"context"
	"testing"
	"time"

	crud "github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestModel represents a test entity for CRUD operations
type TestModel struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique"`
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database")

	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "Failed to migrate test models")

	return db
}

func TestNewRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)

	assert.NotNil(t, repo, "NewRepository should not return nil")
}

func TestRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	t.Run("successful insert", func(t *testing.T) {
		model := TestModel{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}

		result, err := repo.Insert(ctx, model)
		assert.NoError(t, err, "Insert should not return error")
		assert.NotZero(t, result.ID, "Insert should set ID after successful insert")
		assert.Equal(t, model.Name, result.Name, "Insert should preserve Name field")
		assert.Equal(t, model.Email, result.Email, "Insert should preserve Email field")
		assert.Equal(t, model.Age, result.Age, "Insert should preserve Age field")
	})

	t.Run("insert zero value", func(t *testing.T) {
		_, err := repo.Insert(ctx, TestModel{})
		assert.Error(t, err, "Insert should return error for zero value")
		assert.Contains(t, err.Error(), "zero value", "Error should mention zero value")
	})
}

func TestRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	// Insert test data
	testModels := []TestModel{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, model := range testModels {
		_, err := repo.Insert(ctx, model)
		assert.NoError(t, err, "Failed to insert test data")
	}

	t.Run("find all records", func(t *testing.T) {
		results, err := repo.FindAll(ctx, crud.Specification[TestModel]{})
		assert.NoError(t, err, "FindAll should not return error")
		assert.Len(t, results, 3, "FindAll should return all 3 records")
	})

	t.Run("find by name", func(t *testing.T) {
		spec := crud.Specification[TestModel]{
			Model: TestModel{Name: "Alice"},
		}
		results, err := repo.FindAll(ctx, spec)
		assert.NoError(t, err, "FindAll should not return error")
		assert.Len(t, results, 1, "FindAll should return 1 record for Alice")
		assert.Equal(t, "Alice", results[0].Name, "Found record should be Alice")
	})
}

func TestRepository_FindFirst(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	testModel := TestModel{Name: "Alice", Email: "alice@example.com", Age: 25}
	_, err := repo.Insert(ctx, testModel)
	assert.NoError(t, err, "Failed to insert test data")

	t.Run("find existing record", func(t *testing.T) {
		spec := crud.Specification[TestModel]{
			Model: TestModel{Name: "Alice"},
		}
		result, err := repo.FindFirst(ctx, spec)
		assert.NoError(t, err, "FindFirst should not return error")
		assert.NotZero(t, result.ID, "FindFirst should return record with ID when found")
		assert.Equal(t, "Alice", result.Name, "Found record should have correct name")
	})

	t.Run("find non-existent record", func(t *testing.T) {
		spec := crud.Specification[TestModel]{
			Model: TestModel{Name: "NonExistent"},
		}
		result, err := repo.FindFirst(ctx, spec)
		assert.NoError(t, err, "FindFirst should not return error for non-existent record")
		assert.Zero(t, result.ID, "FindFirst should return zero value when not found")
	})
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	testModel := TestModel{Name: "Alice", Email: "alice@example.com", Age: 25}
	inserted, err := repo.Insert(ctx, testModel)
	assert.NoError(t, err, "Failed to insert test data")

	t.Run("successful update", func(t *testing.T) {
		updated := TestModel{
			ID:    inserted.ID,
			Name:  "Alice Updated",
			Email: "alice.updated@example.com",
			Age:   26,
		}

		result, err := repo.Update(ctx, updated)
		assert.NoError(t, err, "Update should not return error")
		assert.Equal(t, updated.Name, result.Name, "Update should change Name field")
		assert.Equal(t, updated.Email, result.Email, "Update should change Email field")
		assert.Equal(t, updated.Age, result.Age, "Update should change Age field")
	})

	t.Run("update zero value", func(t *testing.T) {
		_, err := repo.Update(ctx, TestModel{})
		assert.Error(t, err, "Update should return error for zero value")
		assert.Contains(t, err.Error(), "zero value", "Error should mention zero value")
	})
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	testModel := TestModel{Name: "Alice", Email: "alice@example.com", Age: 25}
	inserted, err := repo.Insert(ctx, testModel)
	assert.NoError(t, err, "Failed to insert test data")

	t.Run("successful delete", func(t *testing.T) {
		err := repo.Delete(ctx, inserted)
		assert.NoError(t, err, "Delete should not return error")

		// Verify record is deleted
		result, err := repo.FindFirst(ctx, crud.Specification[TestModel]{
			Model: TestModel{ID: inserted.ID},
		})
		assert.NoError(t, err, "Error verifying deletion")
		assert.Zero(t, result.ID, "Delete should remove the record")
	})

	t.Run("delete zero value", func(t *testing.T) {
		err := repo.Delete(ctx, TestModel{})
		assert.Error(t, err, "Delete should return error for zero value")
		assert.Contains(t, err.Error(), "zero value", "Error should mention zero value")
	})
}

func TestRepository_InsertMany(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	t.Run("successful batch insert", func(t *testing.T) {
		models := []TestModel{
			{Name: "Alice", Email: "alice@example.com", Age: 25},
			{Name: "Bob", Email: "bob@example.com", Age: 30},
			{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		}

		results, err := repo.InsertMany(ctx, models)
		assert.NoError(t, err, "InsertMany should not return error")
		assert.Len(t, results, len(models), "InsertMany should return same number of records")

		for i, result := range results {
			assert.NotZero(t, result.ID, "InsertMany result[%d] should have ID set", i)
			assert.Equal(t, models[i].Name, result.Name, "InsertMany result[%d] should preserve Name", i)
			assert.Equal(t, models[i].Email, result.Email, "InsertMany result[%d] should preserve Email", i)
			assert.Equal(t, models[i].Age, result.Age, "InsertMany result[%d] should preserve Age", i)
		}
	})

	t.Run("empty batch insert", func(t *testing.T) {
		_, err := repo.InsertMany(ctx, []TestModel{})
		assert.Error(t, err, "InsertMany should return error for empty slice")
		assert.Contains(t, err.Error(), "empty", "Error should mention empty")
	})
}

func TestRepository_GetGormInstance(t *testing.T) {
	db := setupTestDB(t)
	repo := crud.NewRepository[TestModel](db)
	ctx := context.Background()

	instance, err := repo.GetGormInstance(ctx)
	assert.NoError(t, err, "GetGormInstance should not return error")
	assert.NotNil(t, instance, "GetGormInstance should not return nil instance")
}

func TestSpecification(t *testing.T) {
	spec := crud.Specification[TestModel]{
		Model: TestModel{
			Name: "Alice",
			Age:  25,
		},
		PreloadRelations: []string{"Profile", "Posts"},
		ForUpdate:        true,
	}

	assert.Equal(t, "Alice", spec.Model.Name, "Specification should preserve Model.Name")
	assert.Equal(t, 25, spec.Model.Age, "Specification should preserve Model.Age")
	assert.Len(t, spec.PreloadRelations, 2, "Specification should have 2 PreloadRelations")
	assert.Contains(t, spec.PreloadRelations, "Profile", "Specification should contain Profile relation")
	assert.Contains(t, spec.PreloadRelations, "Posts", "Specification should contain Posts relation")
	assert.True(t, spec.ForUpdate, "Specification.ForUpdate should be true")
}
