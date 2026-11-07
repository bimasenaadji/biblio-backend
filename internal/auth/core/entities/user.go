package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an authenticated identity accessing the Biblio platform.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string         `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // Anti-leak: NEVER expose in JSON
	FullName     string         `gorm:"type:varchar(150);not null" json:"full_name"`
	RoleID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"role_id"`
	Role         *Role          `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"role,omitempty"`
	IsActive     bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate generates a UUID if not provided.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
