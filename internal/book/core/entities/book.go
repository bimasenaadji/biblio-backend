package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Book represents the core domain model and PostgreSQL database entity for book titles.
type Book struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ISBN          string         `gorm:"type:varchar(30);uniqueIndex;not null" json:"isbn"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Author        string         `gorm:"type:varchar(150);not null" json:"author"`
	Category      string         `gorm:"type:varchar(100);not null" json:"category"`
	PublishedYear int            `gorm:"not null" json:"published_year"`
	CoverURL      string         `gorm:"type:text" json:"cover_url"`
	CreatedBy     *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	UpdatedBy     *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	Copies        []BookCopy     `gorm:"foreignKey:BookID" json:"copies,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate is a GORM hook that auto-generates a UUID if not provided.
func (b *Book) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
