package ports

import (
	bookEntities "biblio/internal/book/core/entities"
	"biblio/internal/loan/core/entities"
	"context"

	"github.com/google/uuid"
)

// ReturnedCopyUpdate pairs a book copy UUID with its new inventory status based on returned condition.
type ReturnCopyUpdate struct {
	CopyID uuid.UUID
	Status bookEntities.CopyStatus
}

// LoanRepository defines the outbound secondary SPI port for loan data persistence.
// Adheres to Hexagonal Architecture: domain core depends only on this abstraction, not database drivers.
type LoanRepository interface {
	// CreateLoanTransaction coordinates an atomic database transaction that inserts loan headers,
	// child loan items, and transitions all involved book copies to 'borrowed' status.
	CreateLoanTransaction(ctx context.Context, loan *entities.Loan, copyIDs []uuid.UUID) error

	// FindByID retrieves single loan record with nested preloaded associations (Loan -> Items -> Copy -> Book).
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Loan, error)

	// FindByLoanNumber retrieves single loan record by human-readable reference number (e.g., LN-20261105-XXXX).
	FindByLoanNumber(ctx context.Context, loanNumber string) (*entities.Loan, error)

	// ReturnLoanTransaction atomically updates the loan header, item return dates/conditions,
	// and adjusts each physical book copy's availability status.
	ReturnLoanTransaction(ctx context.Context, loan *entities.Loan, copyUpdates []ReturnCopyUpdate) error
}
