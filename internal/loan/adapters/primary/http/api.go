package http

import (
	"biblio/internal/loan/core/entities"
	"biblio/internal/loan/ports"
	"context"

	"github.com/google/uuid"
)

// LoanAPI coordinates mapping between HTTP DTOs and internal domain usecases.
type LoanAPI struct {
	service ports.LoanService
}

// NewLoanAPI instantiates the LoanAPI coordinator.
func NewLoanAPI(service ports.LoanService) *LoanAPI {
	return &LoanAPI{
		service: service,
	}
}

// toLoanResponse converts a domain Loan entity (with preloaded associations) into an outbound LoanResponse DTO.
func toLoanResponse(loan *entities.Loan) *LoanResponse {
	if loan == nil {
		return nil
	}

	// Map child LoanItems into LoanItemResponse slice
	items := make([]LoanItemResponse, 0, len(loan.Items))
	for _, item := range loan.Items {
		itemRes := LoanItemResponse{
			ID:                item.ID,
			LoanID:            item.LoanID,
			BookCopyID:        item.BookCopyID,
			ConditionOnBorrow: string(item.ConditionOnBorrow),
			ReturnedAt:        item.ReturnedAt,
		}

		if item.ConditionOnReturn != nil {
			cond := string(*item.ConditionOnReturn)
			itemRes.ConditionOnReturn = &cond
		}

		// Embed book copy and book title if preloaded by repository
		if item.BookCopy != nil {
			summary := &BookCopySummary{
				Barcode:  item.BookCopy.Barcode,
				RackCode: item.BookCopy.RackCode,
			}
			if item.BookCopy.Book != nil {
				summary.BookTitle = item.BookCopy.Book.Title
				summary.Author = item.BookCopy.Book.Author
			}
			itemRes.BookCopy = summary
		}

		items = append(items, itemRes)
	}

	return &LoanResponse{
		ID:         loan.ID,
		LoanNumber: loan.LoanNumber,
		MemberID:   loan.MemberID,
		MemberName: loan.MemberName,
		Status:     string(loan.Status),
		BorrowedAt: loan.BorrowedAt,
		DueDate:    loan.DueDate,
		ReturnedAt: loan.ReturnedAt,
		Metadata:   loan.Metadata,
		Items:      items,
		TotalItems: len(items),
		CreatedAt:  loan.CreatedAt,
		UpdatedAt:  loan.UpdatedAt,
	}
}

// BorrowBooks handles translation from HTTP request DTO to usecase and maps back to HTTP response DTO.
func (a *LoanAPI) BorrowBooks(ctx context.Context, req BorrowBookRequest) (*LoanResponse, error) {
	input := ports.BorrowRequestInput{
		MemberID:     req.MemberID,
		MemberName:   req.MemberName,
		Barcodes:     req.Barcodes,
		DurationDays: req.DurationDays,
		Metadata:     req.Metadata,
	}

	loan, err := a.service.BorrowBooks(ctx, input)
	if err != nil {
		return nil, err
	}

	return toLoanResponse(loan), nil
}

// GetLoanByID retrieves single loan by UUID and returns its HTTP response DTO.
func (a *LoanAPI) GetLoanByID(ctx context.Context, id uuid.UUID) (*LoanResponse, error) {
	loan, err := a.service.GetLoanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toLoanResponse(loan), nil
}

// GetLoanByNumber retrieves single loan by reference number and returns its HTTP response DTO.
func (a *LoanAPI) GetLoanByNumber(ctx context.Context, loanNumber string) (*LoanResponse, error) {
	loan, err := a.service.GetLoanByNumber(ctx, loanNumber)
	if err != nil {
		return nil, err
	}

	return toLoanResponse(loan), nil
}
