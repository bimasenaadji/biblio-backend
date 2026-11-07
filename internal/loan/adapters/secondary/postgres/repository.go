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

// FindAll retrieves paginated and filtered loan records using a high-performance CTE query.
func (r *loanRepository) FindAll(ctx context.Context, filter ports.LoanFilter) ([]entities.Loan, int64, error) {
	var loans []entities.Loan
	var totalRows int64

	// Base CTE definition: Common Table Expressions (CTE)
	// Solves N+1 query problem by pre-aggregating item stats in a single database round-trip.
	// - COUNT(*): Total count of book copies borrowed in this loan.
	// - BOOL_OR(...): Evaluates to true if AT LEAST ONE book copy was returned in 'damaged' condition.
	// - COALESCE(..., false): Guards against NULL when items have not been returned yet (condition_on_return is NULL).
	cteQuery := `
		WITH item_stats AS (
			SELECT 
				loan_id,
				COUNT(*) AS total_items,
				COALESCE(BOOL_OR(condition_on_return = 'damaged'), false) AS has_damaged_book
			FROM loan_items
			GROUP BY loan_id
		)
	`

	// Step 1: Count total matching records for pagination metadata
	countSQL := "SELECT count(*) FROM loans WHERE 1=1"
	var countArgs []interface{}

	if filter.Status != "" {
		switch filter.Status {
		case "overdue":
			countSQL += " AND status = ? AND due_date < ?"
			countArgs = append(countArgs, entities.LoanStatusBorrowed, time.Now())
		case "borrowed":
			countSQL += " AND status = ?"
			countArgs = append(countArgs, entities.LoanStatusBorrowed)
		case "returned":
			countSQL += " AND status = ?"
			countArgs = append(countArgs, entities.LoanStatusReturned)
		}
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		countSQL += " AND (member_name ILIKE ? OR loan_number ILIKE ?)"
		countArgs = append(countArgs, searchPattern, searchPattern)
	}

	if err := r.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count matching loan records: %w", err)
	}

	// Step 2: Assemble CTE query with dynamic filters and pagination
	// - LEFT JOIN: Guarantees loan headers are retrieved even if they have 0 items.
	// - COALESCE: Guarantees clean zero/false values rather than SQL NULL, preventing unmarshaling panics.
	dataSQL := cteQuery + `
		SELECT 
			loans.*,
			COALESCE(item_stats.total_items, 0) AS total_items,
			COALESCE(item_stats.has_damaged_book, false) AS has_damaged_book
		FROM loans
		LEFT JOIN item_stats ON loans.id = item_stats.loan_id
		WHERE 1=1
	`
	var dataArgs []interface{}

	if filter.Status != "" {
		switch filter.Status {
		case "overdue":
			dataSQL += " AND loans.status = ? AND loans.due_date < ?"
			dataArgs = append(dataArgs, entities.LoanStatusBorrowed, time.Now())
		case "borrowed":
			dataSQL += " AND loans.status = ?"
			dataArgs = append(dataArgs, entities.LoanStatusBorrowed)
		case "returned":
			dataSQL += " AND loans.status = ?"
			dataArgs = append(dataArgs, entities.LoanStatusReturned)
		}
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		dataSQL += " AND (loans.member_name ILIKE ? OR loans.loan_number ILIKE ?)"
		dataArgs = append(dataArgs, searchPattern, searchPattern)
	}

	// Step 3: Apply safe sorting and pagination LIMIT / OFFSET
	offset := (filter.Page - 1) * filter.Limit
	dataSQL += " ORDER BY loans.created_at DESC LIMIT ? OFFSET ?"
	dataArgs = append(dataArgs, filter.Limit, offset)

	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&loans).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch paginated loans via CTE: %w", err)
	}

	return loans, totalRows, nil
}
