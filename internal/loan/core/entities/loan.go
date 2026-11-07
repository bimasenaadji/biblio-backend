package entities

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoanStatus represents the lifecycle state of a book loan transaction.
type LoanStatus string

const (
	LoanStatusBooked   LoanStatus = "booked"
	LoanStatusBorrowed LoanStatus = "borrowed"
	LoanStatusReturned LoanStatus = "returned"
	LoanStatusOverdue  LoanStatus = "overdue"
)

// Loan represents the parent transaction header entity for library book borrowing.
type Loan struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	LoanNumber string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"loan_number"`
	MemberID   *uuid.UUID      `gorm:"type:uuid;index" json:"member_id"`
	MemberName string          `gorm:"type:varchar(150);not null" json:"member_name"`
	Status     LoanStatus      `gorm:"type:varchar(20);not null;default:'borrowed'" json:"status"`
	BorrowedAt time.Time       `gorm:"not null" json:"borrowed_at"`
	DueDate    time.Time       `gorm:"not null" json:"due_date"`
	ReturnedAt *time.Time      `json:"returned_at"`
	Metadata   json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"` // PostgreSQL JSONB dynamic attributes
	CreatedBy  *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy  *uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`

	// Has Many Relation: 1 Loan transaction contains multiple loaned book copy items
	Items []LoanItem `gorm:"foreignKey:LoanID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"items,omitempty"`

	// Virtual / Read-Only Fields (Populated via Raw SQL CTE in repository queries)
	// Tag 'gorm:"->"' marks these fields as read-only: GORM will scan values into them,
	// but AutoMigrate will NOT generate physical database columns for them.
	TotalItems     int  `gorm:"->" json:"total_items,omitempty"`
	HasDamagedBook bool `gorm:"->" json:"has_damaged_book,omitempty"`
}

// BeforeCreate is a GORM lifecycle hook that generates a new UUID and formatted LoanNumber.
func (l *Loan) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}

	// Auto-generate reference number formatted as: LN-YYYYMMDD-XXXX
	if l.LoanNumber == "" {
		dateStr := time.Now().Format("20060102")
		randomDigits := rand.Intn(9000) + 1000
		l.LoanNumber = fmt.Sprintf("LN-%s-%04d", dateStr, randomDigits)
	}
	return nil
}
