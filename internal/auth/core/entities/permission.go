package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents an atomic, granular authorization capability in the system.
// Standard naming convention: "resource:action" (e.g., "book:write", "loan:return").
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
