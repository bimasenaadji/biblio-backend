package usecases

import (
	bookPorts "biblio/internal/book/ports"
	"biblio/internal/loan/core/entities"
	"biblio/internal/loan/ports"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// loanService coordinates business workflows for book loans.
type loanService struct {
	loanRepo ports.LoanRepository
	bookRepo bookPorts.BookRepository
}

// NewLoanService creates a new instance of the loan usecase service.
// Notice it takes bookRepo as an interface dependency to resolve barcodes cleanly!
func NewLoanService(loanRepo ports.LoanRepository, bookRepo bookPorts.BookRepository) ports.LoanService {
	return &loanService{
		loanRepo: loanRepo,
		bookRepo: bookRepo,
	}
}

// BorrowBooks coordinates barcode verification, loan header creation, and atomic transaction execution.
func (s *loanService) BorrowBooks(ctx context.Context, input ports.BorrowRequestInput) (*entities.Loan, error) {
	// Business Rule 1: Validate cart boundaries (must borrow between 1 and 5 books)
	if len(input.Barcodes) == 0 {
		return nil, errors.New("at least one book copy barcode must be provided")
	}
	if len(input.Barcodes) > 5 {
		return nil, errors.New("maximum loan limit exceeded: cannot borrow more than 5 books per transaction")
	}

	// Business Rule 2: Prevent duplicate scans in the same loan request
	scannedSet := make(map[string]bool)
	copyIDs := make([]uuid.UUID, 0, len(input.Barcodes))
	loanItems := make([]entities.LoanItem, 0, len(input.Barcodes))

	// Business Rule 3: Resolve each scanned barcode to its physical BookCopy entity
	for _, barcode := range input.Barcodes {
		if scannedSet[barcode] {
			return nil, fmt.Errorf("duplicate barcode detected in request: %s", barcode)
		}
		scannedSet[barcode] = true

		copy, err := s.bookRepo.FindCopyByBarcode(ctx, barcode)
		if err != nil {
			return nil, fmt.Errorf("book copy with barcode '%s' not found", barcode)
		}

		copyIDs = append(copyIDs, copy.ID)
		loanItems = append(loanItems, entities.LoanItem{
			BookCopyID:        copy.ID,
			ConditionOnBorrow: copy.Condition,
		})
	}

	// Business Rule 4: Compute loan period and due date (default to 7 days if omitted)
	duration := input.DurationDays
	if duration <= 0 {
		duration = 7
	}

	borrowedAt := time.Now()
	dueDate := borrowedAt.AddDate(0, 0, duration)

	// Step 5: Construct the parent Loan entity
	loan := &entities.Loan{
		MemberID:   input.MemberID,
		MemberName: input.MemberName,
		Status:     entities.LoanStatusBorrowed,
		BorrowedAt: borrowedAt,
		DueDate:    dueDate,
		Metadata:   input.Metadata,
		Items:      loanItems,
	}

	// Step 6: Delegate execution to atomic database transaction
	if err := s.loanRepo.CreateLoanTransaction(ctx, loan, copyIDs); err != nil {
		return nil, err
	}

	// Step 7: Return the freshly persisted loan with full preloaded associations
	return s.loanRepo.FindByID(ctx, loan.ID)
}

// GetLoanByID retrieves single loan transaction details by UUID.
func (s *loanService) GetLoanByID(ctx context.Context, id uuid.UUID) (*entities.Loan, error) {
	return s.loanRepo.FindByID(ctx, id)
}

// GetLoanByNumber retrieves single loan transaction details by reference number.
func (s *loanService) GetLoanByNumber(ctx context.Context, loanNumber string) (*entities.Loan, error) {
	return s.loanRepo.FindByLoanNumber(ctx, loanNumber)
}
