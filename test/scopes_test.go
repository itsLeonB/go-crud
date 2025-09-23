package gocrud_test

import (
	"testing"
	"time"

	crud "github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SoftDeleteModel for testing soft delete functionality
type SoftDeleteModel struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func setupScopesTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database")

	err = db.AutoMigrate(&TestModel{}, &SoftDeleteModel{})
	assert.NoError(t, err, "Failed to migrate test models")

	return db
}

func TestPaginate(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	testModels := []TestModel{
		{Name: "User1", Email: "user1@example.com", Age: 25},
		{Name: "User2", Email: "user2@example.com", Age: 30},
		{Name: "User3", Email: "user3@example.com", Age: 35},
		{Name: "User4", Email: "user4@example.com", Age: 40},
		{Name: "User5", Email: "user5@example.com", Age: 45},
	}
	err := db.Create(&testModels).Error
	assert.NoError(t, err, "Failed to create test data")

	tests := []struct {
		name      string
		page      int
		limit     int
		wantCount int
	}{
		{"first page with limit 2", 1, 2, 2},
		{"second page with limit 2", 2, 2, 2},
		{"third page with limit 2", 3, 2, 1},
		{"page 0 should default to page 1", 0, 2, 2},
		{"negative page should default to page 1", -1, 2, 2},
		{"large limit", 1, 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestModel
			err := db.Scopes(crud.Paginate(tt.page, tt.limit)).Find(&results).Error
			assert.NoError(t, err, "Paginate should not return error")
			assert.Len(t, results, tt.wantCount, "Paginate should return expected number of records")
		})
	}
}

func TestOrderBy(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	testModels := []TestModel{
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
	}
	err := db.Create(&testModels).Error
	assert.NoError(t, err, "Failed to create test data")

	tests := []struct {
		name      string
		field     string
		ascending bool
		wantFirst string
		wantErr   bool
	}{
		{"order by name ascending", "name", true, "Alice", false},
		{"order by name descending", "name", false, "Charlie", false},
		{"order by age ascending", "age", true, "Alice", false},
		{"order by age descending", "age", false, "Charlie", false},
		{"invalid field name", "name'; DROP TABLE users; --", true, "", true},
		{"valid field with underscore", "created_at", true, "Charlie", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestModel
			query := db.Scopes(crud.OrderBy(tt.field, tt.ascending))
			err := query.Find(&results).Error

			if tt.wantErr {
				assert.Error(t, err, "OrderBy should return error for invalid field")
				return
			}

			assert.NoError(t, err, "OrderBy should not return error")
			assert.NotEmpty(t, results, "OrderBy should return results")

			if len(results) > 0 {
				assert.Equal(t, tt.wantFirst, results[0].Name, "OrderBy should return correct first result")
			}
		})
	}
}

func TestWhereBySpec(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	testModels := []TestModel{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 25},
	}
	err := db.Create(&testModels).Error
	assert.NoError(t, err, "Failed to create test data")

	tests := []struct {
		name      string
		spec      TestModel
		wantCount int
	}{
		{"find by name", TestModel{Name: "Alice"}, 1},
		{"find by age", TestModel{Age: 25}, 2},
		{"find by name and age", TestModel{Name: "Alice", Age: 25}, 1},
		{"find non-existent", TestModel{Name: "NonExistent"}, 0},
		{"empty spec (find all)", TestModel{}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestModel
			err := db.Scopes(crud.WhereBySpec(tt.spec)).Find(&results).Error
			assert.NoError(t, err, "WhereBySpec should not return error")
			assert.Len(t, results, tt.wantCount, "WhereBySpec should return expected number of records")
		})
	}
}

func TestPreloadRelations(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	user := TestModel{Name: "Alice", Email: "alice@example.com", Age: 25}
	err := db.Create(&user).Error
	assert.NoError(t, err, "Failed to create test data")

	tests := []struct {
		name      string
		relations []string
		wantErr   bool
	}{
		{"preload empty relations", []string{}, false},
		{"preload nil relations", nil, false},
		// Skip testing actual relations since TestModel doesn't have them
		// This tests the scope function itself
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestModel
			err := db.Scopes(crud.PreloadRelations(tt.relations)).First(&result, user.ID).Error

			if tt.wantErr {
				assert.Error(t, err, "PreloadRelations should return error")
				return
			}

			assert.NoError(t, err, "PreloadRelations should not return error")
			assert.NotZero(t, result.ID, "PreloadRelations should return record with ID")
		})
	}
}

func TestBetweenTime(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data with different timestamps
	now := time.Now()
	testModels := []TestModel{
		{Name: "Old", Email: "old@example.com", Age: 25, CreatedAt: now.Add(-2 * time.Hour)},
		{Name: "Recent", Email: "recent@example.com", Age: 30, CreatedAt: now.Add(-1 * time.Hour)},
		{Name: "New", Email: "new@example.com", Age: 35, CreatedAt: now},
	}

	for _, model := range testModels {
		err := db.Create(&model).Error
		assert.NoError(t, err, "Failed to create test data")
	}

	tests := []struct {
		name      string
		col       string
		start     time.Time
		end       time.Time
		wantCount int
	}{
		{"between specific times", "created_at", now.Add(-90 * time.Minute), now.Add(-30 * time.Minute), 1},
		{"from start time only", "created_at", now.Add(-90 * time.Minute), time.Time{}, 2},
		{"until end time only", "created_at", time.Time{}, now.Add(-30 * time.Minute), 2},
		{"both times zero", "created_at", time.Time{}, time.Time{}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestModel
			err := db.Scopes(crud.BetweenTime(tt.col, tt.start, tt.end)).Find(&results).Error
			assert.NoError(t, err, "BetweenTime should not return error")
			assert.Len(t, results, tt.wantCount, "BetweenTime should return expected number of records")
		})
	}
}

func TestDefaultOrder(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data with different timestamps
	now := time.Now()
	testModels := []TestModel{
		{Name: "First", Email: "first@example.com", Age: 25, CreatedAt: now.Add(-2 * time.Hour)},
		{Name: "Second", Email: "second@example.com", Age: 30, CreatedAt: now.Add(-1 * time.Hour)},
		{Name: "Third", Email: "third@example.com", Age: 35, CreatedAt: now},
	}

	for _, model := range testModels {
		err := db.Create(&model).Error
		assert.NoError(t, err, "Failed to create test data")
	}

	t.Run("default order by created_at DESC", func(t *testing.T) {
		var results []TestModel
		err := db.Scopes(crud.DefaultOrder()).Find(&results).Error
		assert.NoError(t, err, "DefaultOrder should not return error")
		assert.Len(t, results, 3, "DefaultOrder should return all 3 records")

		// Should be ordered by created_at DESC, so "Third" should be first
		assert.Equal(t, "Third", results[0].Name, "DefaultOrder should return most recent record first")
		assert.Equal(t, "First", results[2].Name, "DefaultOrder should return oldest record last")
	})
}

func TestForUpdate(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	testModel := TestModel{Name: "Alice", Email: "alice@example.com", Age: 25}
	err := db.Create(&testModel).Error
	assert.NoError(t, err, "Failed to create test data")

	tests := []struct {
		name   string
		enable bool
	}{
		{"for update enabled", true},
		{"for update disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestModel
			err := db.Scopes(crud.ForUpdate(tt.enable)).First(&result, testModel.ID).Error
			assert.NoError(t, err, "ForUpdate should not return error")
			assert.NotZero(t, result.ID, "ForUpdate should return record with ID")
			assert.Equal(t, testModel.Name, result.Name, "ForUpdate should return correct record")
		})
	}
}
func TestDeletedFilter(t *testing.T) {
	db := setupScopesTestDB(t)

	// Insert test data
	models := []SoftDeleteModel{
		{Name: "Active1"},
		{Name: "Active2"},
		{Name: "ToDelete1"},
		{Name: "ToDelete2"},
	}
	err := db.Create(&models).Error
	assert.NoError(t, err, "Failed to create test data")

	// Soft delete some records
	err = db.Delete(&models[2]).Error
	assert.NoError(t, err, "Failed to soft delete record")
	err = db.Delete(&models[3]).Error
	assert.NoError(t, err, "Failed to soft delete record")

	tests := []struct {
		name      string
		filter    crud.DeletedFilter
		useUnscoped bool
		wantCount int
	}{
		{"exclude deleted", crud.ExcludeDeleted, false, 2},
		{"include deleted", crud.IncludeDeleted, true, 4},
		{"only deleted", crud.OnlyDeleted, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []SoftDeleteModel
			query := db.Scopes(tt.filter.WhereDeleted())
			if tt.useUnscoped {
				query = query.Unscoped()
			}
			err := query.Find(&results).Error
			assert.NoError(t, err, "DeletedFilter should not return error")
			assert.Len(t, results, tt.wantCount, "DeletedFilter should return expected number of records")
		})
	}
}
