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
