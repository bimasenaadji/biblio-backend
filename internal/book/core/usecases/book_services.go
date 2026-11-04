package usecases

import (
	"biblio/internal/book/core/entities"
	"biblio/internal/book/ports"
	"context"

	"github.com/google/uuid"
)

// bookService coordinates business logic for book operations.
type bookService struct {
	repo ports.BookRepository
}

// NewBookService creates a new instance of the book use case service.
func NewBookService(repo ports.BookRepository) ports.BookService {
	return &bookService{
		repo: repo,
	}
}

// CreateBook delegates creation to the repository layer.
func (s *bookService) CreateBook(ctx context.Context, book *entities.Book) (*entities.Book, error) {
	if err := s.repo.Create(ctx, book); err != nil {
		return nil, err
	}
	return book, nil
}

// GetAllBooks retrieves all book entities.
func (s *bookService) GetAllBooks(ctx context.Context) ([]entities.Book, error) {
	return s.repo.FindAll(ctx)
}

// GetBookByID finds a single book by UUID.
func (s *bookService) GetBookByID(ctx context.Context, id uuid.UUID) (*entities.Book, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateBook validates book existence before updating attributes.
func (s *bookService) UpdateBook(ctx context.Context, id uuid.UUID, input *entities.Book) (*entities.Book, error) {
	// Business Rule: Ensure book exists before attempting update
	existingBook, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update modifiable fields
	existingBook.ISBN = input.ISBN
	existingBook.Title = input.Title
	existingBook.Author = input.Author
	existingBook.Category = input.Category
	existingBook.PublishedYear = input.PublishedYear

	if err := s.repo.Update(ctx, existingBook); err != nil {
		return nil, err
	}

	return existingBook, nil
}

// DeleteBook validates book existence before executing a soft delete.
func (s *bookService) DeleteBook(ctx context.Context, id uuid.UUID) error {
	// Business Rule: Confirm book exists before deletion (allows returning 404)
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}
