package usecases

import (
	bookEntities "biblio/internal/book/core/entities"
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

// ReturnBooks coordinates the book return workflow, physical condition audit, and inventory status updates.
func (s *loanService) ReturnBooks(ctx context.Context, input ports.ReturnRequestInput) (*entities.Loan, error) {
	// Business Rule 1: Validate request input integrity
	if input.LoanID == uuid.Nil {
		return nil, errors.New("loan ID must be provided")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("at least one return item must be provided")
	}

	// Business Rule 2: Fetch current active loan state with preloaded items and book copies
	loan, err := s.loanRepo.FindByID(ctx, input.LoanID)
	if err != nil {
		return nil, fmt.Errorf("loan transaction not found: %w", err)
	}

	// Business Rule 3: Guard against duplicate returns on already completed loans
	if loan.Status == entities.LoanStatusReturned {
		return nil, errors.New("loan transaction has already been completed and all books returned")
	}

	// Index input return items by barcode for fast lookup
	returnMap := make(map[string]ports.ReturnItemInput)
	for _, item := range input.Items {
		returnMap[item.Barcode] = item
	}

	now := time.Now()
	copyUpdates := make([]ports.ReturnCopyUpdate, 0, len(input.Items))

	// Business Rule 4 & 5: Match returned barcodes, audit condition, and determine new inventory status
	for i := range loan.Items {
		item := &loan.Items[i]
		if item.BookCopy == nil {
			continue
		}

		barcode := item.BookCopy.Barcode
		returnInput, found := returnMap[barcode]
		if !found {
			continue // Item not part of this specific return batch
		}

		// Prevent returning an item that was already returned previously
		if item.ReturnedAt != nil {
			return nil, fmt.Errorf("book copy with barcode '%s' has already been returned", barcode)
		}

		// Normalize condition; default to good if omitted
		cond := bookEntities.Condition(returnInput.ConditionOnReturn)
		if cond == "" {
			cond = bookEntities.ConditionGood
		}

		// Update loan item audit fields
		item.ConditionOnReturn = &cond
		item.ReturnedAt = &now

		// Map physical condition to inventory availability status
		var newStatus bookEntities.CopyStatus
		switch cond {
		case bookEntities.ConditionLost:
			newStatus = bookEntities.StatusLost
		case bookEntities.ConditionDamaged:
			newStatus = bookEntities.StatusMaintenance
		default:
			newStatus = bookEntities.StatusAvailable
		}

		copyUpdates = append(copyUpdates, ports.ReturnCopyUpdate{
			CopyID: item.BookCopyID,
			Status: newStatus,
		})

		// Remove from map to track unrecognized barcodes
		delete(returnMap, barcode)
	}

	// If there are still entries in returnMap, client sent barcodes that do not belong to this loan
	if len(returnMap) > 0 {
		var unknownBarcodes []string
		for b := range returnMap {
			unknownBarcodes = append(unknownBarcodes, b)
		}
		return nil, fmt.Errorf("barcodes %v do not belong to loan transaction '%s'", unknownBarcodes, loan.LoanNumber)
	}

	// Business Rule 6: Check if all items in this loan are now returned
	allReturned := true
	for _, item := range loan.Items {
		if item.ReturnedAt == nil {
			allReturned = false
			break
		}
	}

	if allReturned {
		loan.Status = entities.LoanStatusReturned
		loan.ReturnedAt = &now
	}

	// Business Rule 7: Execute atomic database transaction
	if err := s.loanRepo.ReturnLoanTransaction(ctx, loan, copyUpdates); err != nil {
		return nil, fmt.Errorf("failed to process return transaction: %w", err)
	}

	// Return fresh preloaded loan record
	return s.loanRepo.FindByID(ctx, loan.ID)
}
