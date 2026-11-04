package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Condition defines the physical preservation state of a book copy.
type Condition string

const (
	ConditionGood    Condition = "good"    // Pristine or standard library condition
	ConditionDamaged Condition = "damaged" // Worn, torn pages, or water damaged
	ConditionLost    Condition = "lost"    // Marked missing or lost during inventory audit
)

// CopyStatus defines the real-time circulation availability of a book copy.
type CopyStatus string

const (
	StatusAvailable   CopyStatus = "available"   // On the shelf, ready to be borrowed
	StatusBorrowed    CopyStatus = "borrowed"    // Currently checked out by a member
	StatusMaintenance CopyStatus = "maintenance" // Under repair, binding, or cleaning
	StatusLost        CopyStatus = "lost"        // Confirmed unreturned or lost
)

// BookCopy represents a tangible physical copy of a book title with a unique barcode.
type BookCopy struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	BookID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"book_id"`
	Barcode   string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"barcode"`
	RackCode  string         `gorm:"type:varchar(50);not null" json:"rack_code"`
	Condition Condition      `gorm:"type:varchar(20);not null;default:'good'" json:"condition"`
	Status    CopyStatus     `gorm:"type:varchar(20);not null;default:'available'" json:"status"`
	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	UpdatedBy *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Belongs To Relation: Each physical copy belongs to one Book Title.
	// Constraint RESTRICT prevents deleting a Book Title if physical copies still exist.
	Book *Book `gorm:"foreignKey:BookID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"book,omitempty"`
}

// BeforeCreate is a GORM lifecycle hook that auto-assigns a new UUID if none is provided.
func (c *BookCopy) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
