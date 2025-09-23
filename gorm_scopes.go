package crud

import (
	"reflect"
	"time"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud/internal"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Paginate returns a GORM scope that applies pagination to a query.
// It calculates the appropriate offset based on the page number and limit.
// The page parameter is 1-indexed (minimum value of 1).
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}

		offset := (page - 1) * limit

		return db.Limit(limit).Offset(offset)
	}
}

// OrderBy returns a GORM scope that orders query results by the specified field.
// It uses internal.IsValidFieldName to validate the field name and prevent SQL injection.
// Set ascending to true for ascending order, false for descending.
func OrderBy(field string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Basic validation to prevent SQL injection
		// Only allow alphanumeric characters, underscores, and dots for table.column
		if !internal.IsValidFieldName(field) {
			_ = db.AddError(eris.Errorf("invalid field name: %s", field))
			return db
		}

		if ascending {
			return db.Order(field + " ASC")
		}

		return db.Order(field + " DESC")
	}
}

// WhereBySpec returns a GORM scope that applies a WHERE clause based on the provided struct spec.
// Non-zero fields in spec will be used as AND conditions in the query.
func WhereBySpec[T any](spec T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		v := reflect.ValueOf(spec)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return db // nothing to filter
			}
			return db.Where(spec)
		}
		return db.Where(&spec)
	}
}

// PreloadRelations returns a GORM scope that preloads the specified relations.
// It eager loads related data to avoid N+1 query problems.
func PreloadRelations(relations []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, relation := range relations {
			db = db.Preload(relation)
		}

		return db
	}
}

// BetweenTime returns a GORM scope that filters records between two time values.
// It uses GetTimeRangeClause to generate the appropriate SQL WHERE clause.
// Handles open-ended ranges when either start or end time is zero.
func BetweenTime(col string, start, end time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		query, args := ezutil.GetTimeRangeClause(col, start, end)
		if query == "" {
			return db
		}
		return db.Where(query, args...)
	}
}

// DefaultOrder returns a GORM scope that applies default ordering by created_at DESC.
// This provides consistent ordering for queries that don't specify explicit ordering.
// Assumes the model has a created_at field.
func DefaultOrder() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}
}

// ForUpdate returns a GORM scope that conditionally adds FOR UPDATE locking to queries.
// When enable is true, it adds SELECT ... FOR UPDATE to prevent concurrent modifications.
// Used for pessimistic locking in transaction-critical operations.
func ForUpdate(enable bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if enable {
			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
		}
		return db
	}
}

type DeletedFilter struct {
	filterType internal.DeletedFilterType
}

func (df *DeletedFilter) WhereDeleted() func(*gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		if df.filterType == nil {
			return d
		}
		return df.filterType.WhereDeleted()(d)
	}
}

var (
	ExcludeDeleted DeletedFilter = DeletedFilter{internal.ExcludeDeleted{}}
	IncludeDeleted DeletedFilter = DeletedFilter{internal.IncludeDeleted{}}
	OnlyDeleted    DeletedFilter = DeletedFilter{internal.OnlyDeleted{}}
)
