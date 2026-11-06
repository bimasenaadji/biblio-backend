package postgres

import (
	bookEntities "biblio/internal/book/core/entities"
	"biblio/internal/loan/core/entities"
	"biblio/internal/loan/ports"
	"context"
	"fmt"
	"time"

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

// ReturnLoanTransaction atomically updates the loan header, item return dates/conditions,
// and adjusts each physical book copy's availability status.
func (r *loanRepository) ReturnLoanTransaction(ctx context.Context, loan *entities.Loan, copyUpdates []ports.ReturnCopyUpdate) error {
	// Execute within an atomic transaction closure
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Update parent Loan header (Status, ReturnedAt, UpdatedAt)
		// We use Select to explicitly touch only the lifecycle audit columns
		err := tx.Model(loan).
			Select("Status", "ReturnedAt", "UpdatedAt").
			Updates(loan).Error
		if err != nil {
			return fmt.Errorf("failed to update loan header: %w", err)
		}

		// Step 2: Update each child LoanItem's audit trail (ConditionOnReturn, ReturnedAt)
		for _, item := range loan.Items {
			err := tx.Model(&entities.LoanItem{}).
				Where("id = ?", item.ID).
				Select("ConditionOnReturn", "ReturnedAt", "UpdatedAt").
				Updates(item).Error
			if err != nil {
				return fmt.Errorf("failed to update loan item '%s': %w", item.ID, err)
			}
		}

		// Step 3: Reconcile physical inventory status for each affected BookCopy
		// (e.g., 'available' if returned good, 'maintenance' if damaged, 'lost' if missing)
		for _, update := range copyUpdates {
			err := tx.Model(&bookEntities.BookCopy{}).
				Where("id = ?", update.CopyID).Update("status", update.Status).Error
			if err != nil {
				return fmt.Errorf("failed to update book copy status for '%s': %w", update.CopyID, err)
			}
		}

		// Returning nil automatically triggers COMMIT in PostgreSQL
		return nil
	})
}

// FindAll retrieves paginated and filtered loan records along with the total matched record count.
func (r *loanRepository) FindAll(ctx context.Context, filter ports.LoanFilter) ([]entities.Loan, int64, error) {
	var loans []entities.Loan
	var totalRows int64

	// Start building dynamic query on loans table
	query := r.db.WithContext(ctx).Model(&entities.Loan{})

	// Step 1: Dynamic Status Filtering
	switch filter.Status {
	case "overdue":
		// Overdue loans are currently 'borrowed' whose due_date is in the past
		query = query.Where("status = ? AND due_date < ?", entities.LoanStatusBorrowed, time.Now())
	case "borrowed":
		query = query.Where("status = ?", entities.LoanStatusBorrowed)
	case "returned":
		query = query.Where("status = ?", entities.LoanStatusReturned)
	}

	// Step 2: Dynamic Search Keyword (case-insensitive ILIKE for PostgreSQL)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("member_name ILIKE ? OR loan_number ILIKE ?", searchPattern, searchPattern)
	}

	// Step 3: Count total matching rows BEFORE applying pagination limits
	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count matching loan records: %w", err)
	}

	// Step 4: Calculate pagination OFFSET: (Page - 1) * Limit
	offset := (filter.Page - 1) * filter.Limit

	// Step 5: Execute paginated query with nested eager loading
	err := query.
		Limit(filter.Limit).
		Offset(offset).
		Order("created_at DESC").
		Preload("Items.BookCopy.Book").
		Find(&loans).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch paginated loan records: %w", err)
	}

	// Return data slice and total matching count
	return loans, totalRows, nil
}
