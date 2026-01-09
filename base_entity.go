package crud

import (
	"time"

	"github.com/google/uuid"
)

type BaseEntity struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()"`
	CreatedAt time.Time
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (be BaseEntity) IsZero() bool {
	return be.ID == uuid.Nil
}
