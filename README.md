# go-crud

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Test Coverage](https://img.shields.io/badge/Coverage-84.3%25-brightgreen)](./test/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A **generic CRUD repository library** for Go applications using GORM, providing type-safe database operations with transaction support, query scopes, and comprehensive security features.

## 🚀 Features

- **🔒 Type-Safe Operations**: Generic interfaces for compile-time type safety
- **📊 Transaction Management**: Context-aware transactions with nested support
- **🔍 Query Scopes**: Reusable query builders for pagination, filtering, and ordering
- **🛡️ Security First**: SQL injection prevention and field name validation
- **⚡ Performance**: Optimized queries with preloading and batch operations
- **🧪 Well Tested**: 84.3% test coverage with comprehensive test suite
- **📖 Clean API**: Intuitive interfaces following Go best practices

## 📦 Installation

```bash
go get github.com/itsLeonB/go-crud
```

## 🏗️ Architecture

The library is built around three core components:

### 1. **CRUDRepository** - Type-safe database operations

### 2. **Transactor** - Transaction management with context

### 3. **Query Scopes** - Reusable query builders

## 🔧 Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/itsLeonB/go-crud"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Define your model
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"not null"`
    Email     string    `gorm:"unique"`
    Age       int
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    // Initialize GORM
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto-migrate your models
    db.AutoMigrate(&User{})

    // Create repository and transactor
    userRepo := crud.NewCRUDRepository[User](db)
    transactor := crud.NewTransactor(db)

    ctx := context.Background()

    // Your application logic here...
}
```

## 📚 Core Components

### CRUDRepository Interface

The `CRUDRepository` provides type-safe CRUD operations:

```go
type CRUDRepository[T any] interface {
    Insert(ctx context.Context, model T) (T, error)
    FindAll(ctx context.Context, spec Specification[T]) ([]T, error)
    FindFirst(ctx context.Context, spec Specification[T]) (T, error)
    Update(ctx context.Context, model T) (T, error)
    Delete(ctx context.Context, model T) error
    BatchInsert(ctx context.Context, models []T) ([]T, error)
    GetGormInstance(ctx context.Context) (*gorm.DB, error)
}
```

### Specification Pattern

Use specifications to build dynamic queries:

```go
type Specification[T any] struct {
    Model            T        // Model with fields set for WHERE conditions
    PreloadRelations []string // Relations to eager load
    ForUpdate        bool     // Whether to use SELECT ... FOR UPDATE
}
```

## 💡 Usage Examples

### Basic CRUD Operations

```go
ctx := context.Background()
userRepo := crud.NewCRUDRepository[User](db)

// Create a user
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}
createdUser, err := userRepo.Insert(ctx, user)
if err != nil {
    log.Fatal("Failed to create user:", err)
}

// Find users
spec := crud.Specification[User]{
    Model: User{Age: 30}, // Find users with age 30
}
users, err := userRepo.FindAll(ctx, spec)
if err != nil {
    log.Fatal("Failed to find users:", err)
}

// Find first user
firstUser, err := userRepo.FindFirst(ctx, spec)
if err != nil {
    log.Fatal("Failed to find user:", err)
}

// Update user
createdUser.Name = "John Smith"
updatedUser, err := userRepo.Update(ctx, createdUser)
if err != nil {
    log.Fatal("Failed to update user:", err)
}

// Delete user
err = userRepo.Delete(ctx, updatedUser)
if err != nil {
    log.Fatal("Failed to delete user:", err)
}
```

### Batch Operations

```go
// Batch insert multiple users
users := []User{
    {Name: "Alice", Email: "alice@example.com", Age: 25},
    {Name: "Bob", Email: "bob@example.com", Age: 30},
    {Name: "Charlie", Email: "charlie@example.com", Age: 35},
}

createdUsers, err := userRepo.BatchInsert(ctx, users)
if err != nil {
    log.Fatal("Failed to batch insert users:", err)
}
```

### Advanced Queries with Specifications

```go
// Find users with preloaded relations
spec := crud.Specification[User]{
    Model:            User{Age: 25},
    PreloadRelations: []string{"Profile", "Posts"},
    ForUpdate:        false,
}
users, err := userRepo.FindAll(ctx, spec)

// Pessimistic locking
spec = crud.Specification[User]{
    Model:     User{ID: 1},
    ForUpdate: true, // SELECT ... FOR UPDATE
}
user, err := userRepo.FindFirst(ctx, spec)
```

## 🔄 Transaction Management

### Basic Transactions

```go
transactor := crud.NewTransactor(db)

err := transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
    // All operations within this function use the same transaction

    user := User{Name: "Alice", Email: "alice@example.com"}
    createdUser, err := userRepo.Insert(txCtx, user)
    if err != nil {
        return err // Transaction will be rolled back
    }

    // Update in same transaction
    createdUser.Age = 25
    _, err = userRepo.Update(txCtx, createdUser)
    if err != nil {
        return err // Transaction will be rolled back
    }

    return nil // Transaction will be committed
})

if err != nil {
    log.Fatal("Transaction failed:", err)
}
```

### Manual Transaction Control

```go
// Begin transaction
txCtx, err := transactor.Begin(ctx)
if err != nil {
    log.Fatal("Failed to begin transaction:", err)
}

// Perform operations
user, err := userRepo.Insert(txCtx, User{Name: "Bob"})
if err != nil {
    transactor.Rollback(txCtx)
    log.Fatal("Failed to insert user:", err)
}

// Commit transaction
err = transactor.Commit(txCtx)
if err != nil {
    log.Fatal("Failed to commit transaction:", err)
}
```

### Nested Transactions

```go
err := transactor.WithinTransaction(ctx, func(outerTxCtx context.Context) error {
    // Outer transaction
    user, err := userRepo.Insert(outerTxCtx, User{Name: "Outer"})
    if err != nil {
        return err
    }

    // Nested transaction (reuses the same transaction)
    return transactor.WithinTransaction(outerTxCtx, func(innerTxCtx context.Context) error {
        // Inner operations use the same transaction
        return userRepo.Update(innerTxCtx, user)
    })
})
```

## 🔍 Query Scopes

The library provides powerful query scopes for common operations:

### Pagination

```go
db.Scopes(crud.Paginate(page, limit)).Find(&users)

// Example: Get page 2 with 10 items per page
db.Scopes(crud.Paginate(2, 10)).Find(&users)
```

### Ordering

```go
// Order by name ascending
db.Scopes(crud.OrderBy("name", true)).Find(&users)

// Order by created_at descending
db.Scopes(crud.OrderBy("created_at", false)).Find(&users)

// Default ordering (created_at DESC)
db.Scopes(crud.DefaultOrder()).Find(&users)
```

### Filtering

```go
// Filter by specification
spec := User{Age: 25, Name: "Alice"}
db.Scopes(crud.WhereBySpec(spec)).Find(&users)

// Time range filtering
start := time.Now().Add(-24 * time.Hour)
end := time.Now()
db.Scopes(crud.BetweenTime("created_at", start, end)).Find(&users)
```

### Preloading Relations

```go
relations := []string{"Profile", "Posts", "Comments"}
db.Scopes(crud.PreloadRelations(relations)).Find(&users)
```

### Pessimistic Locking

```go
// Add FOR UPDATE clause
db.Scopes(crud.ForUpdate(true)).First(&user, id)
```

### Combining Scopes

```go
// Complex query with multiple scopes
db.Scopes(
    crud.WhereBySpec(User{Age: 25}),
    crud.OrderBy("name", true),
    crud.Paginate(1, 10),
    crud.PreloadRelations([]string{"Profile"}),
).Find(&users)
```

## 🛡️ Security Features

### SQL Injection Prevention

The library includes robust field name validation to prevent SQL injection:

```go
// Safe field names (allowed)
"name"           ✅
"created_at"     ✅
"users.name"     ✅
"table.column"   ✅

// Dangerous field names (rejected)
"name'; DROP TABLE users; --"  ❌
"name' OR '1'='1"              ❌
".invalid"                     ❌
"invalid."                     ❌
"table..column"                ❌
```

### Field Name Validation Rules

- Only ASCII letters, digits, underscores, and dots allowed
- No leading or trailing dots
- No consecutive dots
- No empty strings
- No special characters or SQL keywords

## 🧪 Testing

The library comes with comprehensive tests achieving **84.3% coverage**:

```bash
# Run all tests
go test -v ./test/...

# Run tests with coverage
go test -v -coverprofile=coverage.out -coverpkg=./... ./test/...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Use the test runner script
./test/run_tests.sh
```

### Test Categories

- **Unit Tests**: All public functions and methods
- **Integration Tests**: Database operations with SQLite
- **Security Tests**: SQL injection prevention
- **Concurrency Tests**: Thread safety validation
- **Performance Tests**: Benchmarks for critical functions

## 📁 Project Structure

```
go-crud/
├── gorm_crud_repository.go    # Main CRUD repository implementation
├── gorm_scopes.go            # Query scopes and builders
├── gorm_transactor.go        # Transaction management
├── internal/
│   ├── gorm.go              # Internal GORM utilities
│   └── transactor.go        # Internal transaction logic
├── lib/
│   └── constants.go         # Library constants
└── test/                    # Comprehensive test suite
    ├── crud_repository_test.go
    ├── scopes_test.go
    ├── transactor_test.go
    ├── internal_test.go
    ├── constants_test.go
    └── run_tests.sh
```

## 🔧 Configuration

### Database Drivers

The library works with any GORM-supported database:

```go
// PostgreSQL
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// MySQL
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// SQLite
db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
```

### GORM Configuration

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
    NamingStrategy: schema.NamingStrategy{
        TablePrefix:   "app_",
        SingularTable: false,
    },
})
```

## 🚀 Performance Tips

### 1. Use Batch Operations

```go
// Instead of multiple Insert calls
users := []User{{Name: "Alice"}, {Name: "Bob"}}
userRepo.BatchInsert(ctx, users)
```

### 2. Preload Relations Wisely

```go
// Only preload what you need
spec := crud.Specification[User]{
    PreloadRelations: []string{"Profile"}, // Not all relations
}
```

### 3. Use Pagination

```go
// For large datasets
db.Scopes(crud.Paginate(page, 50)).Find(&users)
```

### 4. Leverage Transactions

```go
// Group related operations
transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
    // Multiple related operations
    return nil
})
```

## 🤝 Contributing

Contributions are welcome! Please ensure:

1. **Tests**: Add tests for new features
2. **Coverage**: Maintain or improve test coverage
3. **Documentation**: Update documentation for new features
4. **Security**: Follow security best practices
5. **Performance**: Consider performance implications

### Development Setup

```bash
# Clone the repository
git clone https://github.com/itsLeonB/go-crud.git
cd go-crud

# Install dependencies
go mod tidy

# Run tests
make test

# Check coverage
make test-coverage-html
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **GORM**: The fantastic ORM library that powers this project
- **Testify**: For making tests more readable and maintainable
- **Eris**: For enhanced error handling and stack traces

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/itsLeonB/go-crud/issues)
- **Discussions**: [GitHub Discussions](https://github.com/itsLeonB/go-crud/discussions)
- **Documentation**: This README and inline code documentation

---

**Built with ❤️ for the Go community**
