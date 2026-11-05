package postgres

import (
	bookEntities "biblio/internal/book/core/entities"
	"biblio/internal/loan/core/entities"
	"biblio/internal/loan/ports"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// loanRepository implements ports.LoanRepository using GORM.
type loanRepository struct {
	db *gorm.DB
}

// NewLoanRepository creates a new instance of the PostgreSQL loan repository.
func NewLoanRepository(db *gorm.DB) ports.LoanRepository {
	return &loanRepository{
		db: db,
	}
}

// CreateLoanTransaction executes an atomic all-or-nothing database transaction.
func (r *loanRepository) CreateLoanTransaction(ctx context.Context, loan *entities.Loan, copyIDs []uuid.UUID) error {
	// Execute within an atomic transaction closure
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Concurrency Guard - Verify that ALL requested book copies exist and are currently 'available'
		var availableCopies []bookEntities.BookCopy
		err := tx.Where("id IN ? AND status = ?", copyIDs, bookEntities.StatusAvailable).Find(&availableCopies).Error
		if err != nil {
			return fmt.Errorf("failed to query copy availability: %w", err)
		}

		// If the count of available copies does not match the requested count, abort immediately!
		if len(availableCopies) != len(copyIDs) {
			return fmt.Errorf("one or more requested book copies are not available for borrowing")
		}

		// Step 2: Insert the parent Loan header record (BeforeCreate hook generates UUID and loan_number)
		if err := tx.Create(loan).Error; err != nil {
			return fmt.Errorf("failed to create loan header: %w", err)
		}

		// Step 3: Assign the generated loan.ID to each loan item and bulk insert them
		for i := range loan.Items {
			loan.Items[i].LoanID = loan.ID
		}
		if err := tx.Create(&loan.Items).Error; err != nil {
			return fmt.Errorf("failed to create loan items: %w", err)
		}

		// Step 4: Transition the status of all borrowed book copies to 'borrowed'
		err = tx.Model(&bookEntities.BookCopy{}).
			Where("id IN ?", copyIDs).
			Update("status", bookEntities.StatusBorrowed).Error
		if err != nil {
			return fmt.Errorf("failed to update book copies status: %w", err)
		}

		// Returning nil signals GORM to automatically COMMIT the transaction
		return nil
	})
}

// FindByID retrieves a loan by its UUID with full nested associations preloaded.
func (r *loanRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Loan, error) {
	var loan entities.Loan
	// Nested Preload: Loan -> Items -> BookCopy -> parent Book Title
	err := r.db.WithContext(ctx).
		Preload("Items.BookCopy.Book").
		First(&loan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

// FindByLoanNumber retrieves a loan by reference number (e.g. LN-20261105-1234) with full associations.
func (r *loanRepository) FindByLoanNumber(ctx context.Context, loanNumber string) (*entities.Loan, error) {
	var loan entities.Loan
	err := r.db.WithContext(ctx).
		Preload("Items.BookCopy.Book").
		First(&loan, "loan_number = ?", loanNumber).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}
