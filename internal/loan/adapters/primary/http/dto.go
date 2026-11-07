package http

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BorrowBookRequest defines the incoming JSON payload for checking out books by barcode.
// It acts as an inbound security barrier to prevent mass assignment vulnerabilities.
// Enforces a strict business boundary: clients can borrow between 1 and 5 books per transaction.
type BorrowBookRequest struct {
	MemberID     *uuid.UUID      `json:"member_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	MemberName   string          `json:"member_name" validate:"required" example:"John Doe"`
	Barcodes     []string        `json:"barcodes" validate:"required,min=1,max=5" example:"[\"BC-CLN-001\",\"BC-CLN-002\"]"`
	DurationDays int             `json:"duration_days" example:"7"`
	Metadata     json.RawMessage `json:"metadata,omitempty" swaggertype:"object"` // PostgreSQL JSONB dynamic attributes
}

// LoanItemResponse defines the outbound JSON representation of an individual borrowed book item.
type LoanItemResponse struct {
	ID                uuid.UUID        `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	LoanID            uuid.UUID        `json:"loan_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BookCopyID        uuid.UUID        `json:"book_copy_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	ConditionOnBorrow string           `json:"condition_on_borrow" example:"good"`
	ConditionOnReturn *string          `json:"condition_on_return,omitempty" example:"good"`
	ReturnedAt        *time.Time       `json:"returned_at,omitempty"`
	BookCopy          *BookCopySummary `json:"book_copy,omitempty"`
}

// BookCopySummary embeds essential physical copy details along with parent book title.
type BookCopySummary struct {
	Barcode   string `json:"barcode" example:"BC-CLN-001"`
	RackCode  string `json:"rack_code" example:"RACK-A1"`
	BookTitle string `json:"book_title,omitempty" example:"Clean Code"`
	Author    string `json:"author,omitempty" example:"Robert C. Martin"`
}

// LoanResponse defines the comprehensive receipt payload delivered to the client.
// Preloaded associations provide the client with rich metadata in a single response envelope.
type LoanResponse struct {
	ID             uuid.UUID          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LoanNumber     string             `json:"loan_number" example:"LN-20261105-0001"`
	MemberID       *uuid.UUID         `json:"member_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	MemberName     string             `json:"member_name" example:"John Doe"`
	Status         string             `json:"status" example:"borrowed"`
	BorrowedAt     time.Time          `json:"borrowed_at"`
	DueDate        time.Time          `json:"due_date"`
	ReturnedAt     *time.Time         `json:"returned_at,omitempty"`
	Metadata       json.RawMessage    `json:"metadata,omitempty" swaggertype:"object"`
	Items          []LoanItemResponse `json:"items,omitempty"`
	// Aggregated/Calculated fields for efficient admin monitoring (avoiding N+1 queries)
	TotalItems     int                `json:"total_items" example:"2"`
	HasDamagedBook bool               `json:"has_damaged_book" example:"false"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ReturnItemRequest represents single returned book copy payload with its audited condition.
type ReturnItemRequest struct {
	Barcode           string `json:"barcode" validate:"required" example:"BC-CLN-001"`
	ConditionOnReturn string `json:"condition_on_return" example:"good"` // "good", "damaged", "lost"
}

// ReturnBooksRequest defines incoming JSON payload for returning borrowed books.
type ReturnBooksRequest struct {
	Items []ReturnItemRequest `json:"items" validate:"required,min=1"`
}
