package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a business persona or organizational job title.
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// BeforeCreate generates a UUID if not provided.
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
