package http

import (
	"biblio/internal/book/core/entities"
	"biblio/internal/book/ports"
	"context"

	"github.com/google/uuid"
)

// BookAPI coordinates mapping between HTTP DTOs and internal domain entities.
type BookAPI struct {
	service ports.BookService
}

// NewBookAPI creates a new instance of the BookAPI coordinator.
func NewBookAPI(service ports.BookService) *BookAPI {
	return &BookAPI{
		service: service,
	}
}

// toBookResponse maps a domain Book entity into an outbound BookResponse DTO.
func toBookResponse(book *entities.Book) *BookResponse {
	return &BookResponse{
		ID:            book.ID,
		ISBN:          book.ISBN,
		Title:         book.Title,
		Author:        book.Author,
		Category:      book.Category,
		PublishedYear: book.PublishedYear,
		CoverURL:      book.CoverURL,
		CreatedAt:     book.CreatedAt,
		UpdatedAt:     book.UpdatedAt,
	}
}

// CreateBook converts request DTO to entity, invokes the service, and returns response DTO.
func (a *BookAPI) CreateBook(ctx context.Context, req CreateBookRequest) (*BookResponse, error) {
	bookEntity := &entities.Book{
		ISBN:          req.ISBN,
		Title:         req.Title,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
	}

	created, err := a.service.CreateBook(ctx, bookEntity)
	if err != nil {
		return nil, err
	}

	return toBookResponse(created), nil
}

// GetAllBooks retrieves all books and maps them into a list of response DTOs.
func (a *BookAPI) GetAllBooks(ctx context.Context) ([]BookResponse, error) {
	books, err := a.service.GetAllBooks(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize empty slice so empty query produces `[]` instead of `null`
	responses := make([]BookResponse, 0, len(books))
	for _, b := range books {
		bookCopy := b
		responses = append(responses, *toBookResponse(&bookCopy))
	}

	return responses, nil
}

// GetBookByID fetches a book by UUID and transforms it into a response DTO.
func (a *BookAPI) GetBookByID(ctx context.Context, id uuid.UUID) (*BookResponse, error) {
	book, err := a.service.GetBookByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toBookResponse(book), nil
}

// UpdateBook maps update request to entity, applies changes, and returns response DTO.
func (a *BookAPI) UpdateBook(ctx context.Context, id uuid.UUID, req UpdateBookRequest) (*BookResponse, error) {
	inputEntity := &entities.Book{
		ISBN:          req.ISBN,
		Title:         req.Title,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
	}

	updated, err := a.service.UpdateBook(ctx, id, inputEntity)
	if err != nil {
		return nil, err
	}

	return toBookResponse(updated), nil
}

// DeleteBook delegates deletion directly to the service layer.
func (a *BookAPI) DeleteBook(ctx context.Context, id uuid.UUID) error {
	return a.service.DeleteBook(ctx, id)
}

// toBookCopyResponse converts a BookCopy entity into an outbound BookCopyResponse DTO.
func toBookCopyResponse(copy *entities.BookCopy) *BookCopyResponse {
	// Step 1: Map copy fields from domain entity to DTO representation
	resp := &BookCopyResponse{
		ID:        copy.ID,
		BookID:    copy.BookID,
		Barcode:   copy.Barcode,
		RackCode:  copy.RackCode,
		Condition: string(copy.Condition),
		Status:    string(copy.Status),
		CreatedAt: copy.CreatedAt,
		UpdatedAt: copy.UpdatedAt,
	}

	// Step 2: If parent book was eagerly joined (e.g. from barcode scan), map and attach it
	if copy.Book != nil {
		resp.Book = toBookResponse(copy.Book)
	}

	return resp
}

// AddBookCopy maps copy request to entity and delegates to the usecase service.
func (a *BookAPI) AddBookCopy(ctx context.Context, bookID uuid.UUID, req CreateBookCopyRequest) (*BookCopyResponse, error) {
	// Step 1: Construct domain entity from incoming request DTO and route parameter bookID
	copyEntity := &entities.BookCopy{
		BookID:    bookID,
		Barcode:   req.Barcode,
		RackCode:  req.RackCode,
		Condition: entities.Condition(req.Condition),
	}

	// Step 2: Invoke the business service (handles parent validation and persistence)
	created, err := a.service.AddBookCopy(ctx, copyEntity)
	if err != nil {
		return nil, err
	}

	// Step 3: Transform persisted domain entity into outbound response DTO
	return toBookCopyResponse(created), nil
}

// GetCopiesByBookID fetches all physical copies of a book title and transforms them into DTOs.
func (a *BookAPI) GetCopiesByBookID(ctx context.Context, bookID uuid.UUID) ([]BookCopyResponse, error) {
	// Step 1: Query copies through the business usecase service
	copies, err := a.service.GetCopiesByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	// Step 2: Pre-allocate slice with exact capacity to avoid re-allocations and prevent null JSON
	responses := make([]BookCopyResponse, 0, len(copies))
	for _, c := range copies {
		copyItem := c
		responses = append(responses, *toBookCopyResponse(&copyItem))
	}

	return responses, nil
}

// GetCopyByBarcode fetches a single physical copy by barcode and maps to response DTO.
func (a *BookAPI) GetCopyByBarcode(ctx context.Context, barcode string) (*BookCopyResponse, error) {
	// Step 1: Lookup copy by barcode via service layer
	copy, err := a.service.GetCopyByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}

	// Step 2: Transform entity (with its preloaded book title) into outbound DTO
	return toBookCopyResponse(copy), nil
}
