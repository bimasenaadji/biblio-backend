package ports

import (
	"biblio/internal/loan/core/entities"
	"context"

	"github.com/google/uuid"
)

// BorrowRequestInput encapsulates the validated business input required to checkout books.
// Decouples HTTP controller payloads from core domain workflow parameters.
type BorrowRequestInput struct {
	MemberID     *uuid.UUID
	MemberName   string
	Barcodes     []string
	DurationDays int
	Metadata     []byte
}

// ReturnItemInput captures the physical condition of each returned book copy.
type ReturnItemInput struct {
	Barcode           string
	ConditionOnReturn string
}

// ReturnRequestInput encapsulates payload for returning books for a loan transaction.
type ReturnRequestInput struct {
	LoanID uuid.UUID
	Items  []ReturnItemInput
}

// LoanService defines the inbound primary API port exposing loan business usecases.
type LoanService interface {
	// BorrowBooks executes the end-to-end checkout workflow including barcode verification and atomic persistence.
	BorrowBooks(ctx context.Context, input BorrowRequestInput) (*entities.Loan, error)

	// GetLoanByID retrieves loan details by internal UUID.
	GetLoanByID(ctx context.Context, id uuid.UUID) (*entities.Loan, error)

	// GetLoanByNumber retrieves loan details by external reference number.
	GetLoanByNumber(ctx context.Context, loanNumber string) (*entities.Loan, error)

	// ReturnBooks coordinates the book return workflow, physical condition audit, and inventory status updates.
	ReturnBooks(ctx context.Context, input ReturnRequestInput) (*entities.Loan, error)
}
