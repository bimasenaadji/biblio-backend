package entities

import (
	bookEntities "biblio/internal/book/core/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoanItem represents the line item for each specific physical book copy within a loan.
// Tracks the physical condition at checkout and return for inventory integrity auditing.
type LoanItem struct {
	ID                uuid.UUID               `gorm:"type:uuid;primaryKey" json:"id"`
	LoanID            uuid.UUID               `gorm:"type:uuid;not null;index" json:"loan_id"`
	BookCopyID        uuid.UUID               `gorm:"type:uuid;not null;index" json:"book_copy_id"`
	ConditionOnBorrow bookEntities.Condition  `gorm:"type:varchar(20);not null;default:'good'" json:"condition_on_borrow"`
	ConditionOnReturn *bookEntities.Condition `gorm:"type:varchar(20)" json:"condition_on_return"`
	ReturnedAt        *time.Time              `json:"returned_at"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`

	// Relational Associations
	// Belongs To: Links to the tangible physical book copy (and through it, parent book metadata)
	BookCopy *bookEntities.BookCopy `gorm:"foreignKey:BookCopyID" json:"book_copy,omitempty"`
}

// BeforeCreate is a GORM lifecycle hook that generates a new UUID for the loan item.
func (item *LoanItem) BeforeCreate(tx *gorm.DB) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return nil
}
