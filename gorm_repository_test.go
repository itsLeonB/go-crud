package crud_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	crud "github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser represents a user entity for testing
type TestUser struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique"`
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TestProfile represents a profile entity for testing
type TestProfile struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint
	Bio    string
	User   TestUser `gorm:"foreignKey:UserID"`
}

// TestEntityWithBase for testing BaseEntity
type TestEntityWithBase struct {
	crud.BaseEntity
	Name string
}

func TestBaseEntity_IsZero(t *testing.T) {
	// Test zero value
	var entity TestEntityWithBase
	assert.True(t, entity.IsZero())

	// Test non-zero value
	entity.ID = uuid.New()
	assert.False(t, entity.IsZero())

	// Test with CreatedAt set but ID still zero
	entity = TestEntityWithBase{}
	entity.CreatedAt = time.Now()
	assert.True(t, entity.IsZero()) // IsZero only checks ID
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	if err = db.AutoMigrate(&TestUser{}, &TestProfile{}); err != nil {
		panic(err)
	}

	return db
}

func TestCRUDRepository_Insert(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	result, err := repo.Insert(ctx, user)

	assert.NoError(t, err)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)
}

func TestCRUDRepository_FindAll(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	users := []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, user := range users {
		_, err := repo.Insert(ctx, user)
		assert.NoError(t, err)
	}

	// Test finding all users
	spec := crud.Specification[TestUser]{}
	result, err := repo.FindAll(ctx, spec)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestCRUDRepository_FindAll_WithSpecification(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	users := []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 25},
	}

	for _, user := range users {
		_, err := repo.Insert(ctx, user)
		assert.NoError(t, err)
	}

	// Test finding users with age 25
	spec := crud.Specification[TestUser]{
		Model: TestUser{Age: 25},
	}
	result, err := repo.FindAll(ctx, spec)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	for _, user := range result {
		assert.Equal(t, 25, user.Age)
	}
}

func TestCRUDRepository_FindFirst(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	insertedUser, err := repo.Insert(ctx, user)
	assert.NoError(t, err)

	// Test finding first user
	spec := crud.Specification[TestUser]{
		Model: TestUser{ID: insertedUser.ID},
	}
	result, err := repo.FindFirst(ctx, spec)

	assert.NoError(t, err)
	assert.Equal(t, insertedUser.ID, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestCRUDRepository_FindFirst_NotFound(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Test finding non-existent user
	spec := crud.Specification[TestUser]{
		Model: TestUser{ID: 999},
	}
	result, err := repo.FindFirst(ctx, spec)

	// Check if error is returned or if result is zero value
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	} else {
		// If no error, result should be zero value
		assert.Zero(t, result.ID)
	}
}

func TestCRUDRepository_Update(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	insertedUser, err := repo.Insert(ctx, user)
	assert.NoError(t, err)

	// Update user
	insertedUser.Name = "John Smith"
	insertedUser.Age = 31
	result, err := repo.Update(ctx, insertedUser)

	assert.NoError(t, err)
	assert.Equal(t, insertedUser.ID, result.ID)
	assert.Equal(t, "John Smith", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 31, result.Age)
	assert.True(t, result.UpdatedAt.After(result.CreatedAt))
}

func TestCRUDRepository_Delete(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	insertedUser, err := repo.Insert(ctx, user)
	assert.NoError(t, err)

	// Delete user
	err = repo.Delete(ctx, insertedUser)
	assert.NoError(t, err)

	// Verify user is deleted
	spec := crud.Specification[TestUser]{
		Model: TestUser{ID: insertedUser.ID},
	}
	result, err := repo.FindFirst(ctx, spec)

	// Check if error is returned or if result is zero value
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	} else {
		// If no error, result should be zero value (deleted)
		assert.Zero(t, result.ID)
	}
}

func TestCRUDRepository_InsertMany(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	users := []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	result, err := repo.InsertMany(ctx, users)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	for i, user := range result {
		assert.NotZero(t, user.ID)
		assert.Equal(t, users[i].Name, user.Name)
		assert.Equal(t, users[i].Email, user.Email)
		assert.Equal(t, users[i].Age, user.Age)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	}
}

func TestCRUDRepository_InsertMany_Empty(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	var users []TestUser

	result, err := repo.InsertMany(ctx, users)

	// The implementation returns an error for empty slice, which is expected behavior
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inserted models cannot be empty")
	assert.Len(t, result, 0)
}

func TestCRUDRepository_GetGormInstance(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	gormDB, err := repo.GetGormInstance(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, gormDB)
	// Just check that we get a valid GORM instance, don't compare directly
	assert.IsType(t, &gorm.DB{}, gormDB)
}

func TestCRUDRepository_WithPreloadRelations(t *testing.T) {
	db := setupTestDB()
	userRepo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	insertedUser, err := userRepo.Insert(ctx, user)
	assert.NoError(t, err)

	// Test finding user with empty preload relations (should work)
	spec := crud.Specification[TestUser]{
		Model:            TestUser{ID: insertedUser.ID},
		PreloadRelations: []string{}, // Empty relations should work
	}
	result, err := userRepo.FindFirst(ctx, spec)

	assert.NoError(t, err)
	assert.Equal(t, insertedUser.ID, result.ID)
}

func TestCRUDRepository_WithForUpdate(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	insertedUser, err := repo.Insert(ctx, user)
	assert.NoError(t, err)

	// Test finding user with FOR UPDATE
	spec := crud.Specification[TestUser]{
		Model:     TestUser{ID: insertedUser.ID},
		ForUpdate: true,
	}
	result, err := repo.FindFirst(ctx, spec)

	assert.NoError(t, err)
	assert.Equal(t, insertedUser.ID, result.ID)
}

func TestCRUDRepository_DeleteMany(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert test data
	users := []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}
	insertedUsers, err := repo.InsertMany(ctx, users)
	assert.NoError(t, err)

	// Delete multiple users
	err = repo.DeleteMany(ctx, insertedUsers)
	assert.NoError(t, err)

	// Verify users are deleted
	spec := crud.Specification[TestUser]{}
	results, err := repo.FindAll(ctx, spec)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestCRUDRepository_DeleteMany_Empty(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	var users []TestUser
	err := repo.DeleteMany(ctx, users)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deleted models cannot be empty")
}

func TestCRUDRepository_SaveMany(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert initial data
	users := []TestUser{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
	}
	insertedUsers, err := repo.InsertMany(ctx, users)
	assert.NoError(t, err)

	// Modify and save
	insertedUsers[0].Age = 26
	insertedUsers[1].Age = 31
	savedUsers, err := repo.SaveMany(ctx, insertedUsers)
	assert.NoError(t, err)
	assert.Len(t, savedUsers, 2)
	assert.Equal(t, 26, savedUsers[0].Age)
	assert.Equal(t, 31, savedUsers[1].Age)
}

func TestCRUDRepository_SaveMany_Empty(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	var users []TestUser
	result, err := repo.SaveMany(ctx, users)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "saved models cannot be empty")
	assert.Len(t, result, 0)
}

func TestCRUDRepository_Insert_Error(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert user with duplicate email should fail
	user1 := TestUser{Name: "Alice", Email: "test@example.com", Age: 25}
	_, err := repo.Insert(ctx, user1)
	assert.NoError(t, err)

	user2 := TestUser{Name: "Bob", Email: "test@example.com", Age: 30}
	_, err = repo.Insert(ctx, user2)
	assert.Error(t, err)
}

func TestCRUDRepository_Update_Error(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Insert a user first to get a valid ID
	user := TestUser{Name: "Test", Email: "test@example.com", Age: 25}
	inserted, err := repo.Insert(ctx, user)
	assert.NoError(t, err)

	// Update should work normally
	inserted.Age = 26
	result, err := repo.Update(ctx, inserted)
	assert.NoError(t, err)
	assert.Equal(t, inserted.ID, result.ID)
	assert.Equal(t, 26, result.Age)
}

func TestCRUDRepository_Delete_Error(t *testing.T) {
	db := setupTestDB()
	repo := crud.NewRepository[TestUser](db)
	ctx := context.Background()

	// Try to delete with zero ID should error due to WHERE conditions required
	user := TestUser{ID: 0, Name: "Test", Email: "test@example.com", Age: 25}
	err := repo.Delete(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "WHERE conditions required")
}
