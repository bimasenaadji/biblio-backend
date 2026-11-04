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

// AddBookCopy validates parent book existence before creating a physical copy.
func (s *bookService) AddBookCopy(ctx context.Context, copy *entities.BookCopy) (*entities.BookCopy, error) {
	// Business Rule 1: Guard against orphan copies — verify parent book title exists in database
	if _, err := s.repo.FindByID(ctx, copy.BookID); err != nil {
		return nil, err
	}

	// Business Rule 2: Set sensible defaults if optional fields are omitted
	if copy.Status == "" {
		copy.Status = entities.StatusAvailable
	}
	if copy.Condition == "" {
		copy.Condition = entities.ConditionGood
	}

	// Business Rule 3: Delegate persistence to the repository layer
	if err := s.repo.CreateCopy(ctx, copy); err != nil {
		return nil, err
	}

	return copy, nil
}

// GetCopiesByBookID verifies parent book existence before retrieving all its physical copies.
func (s *bookService) GetCopiesByBookID(ctx context.Context, bookID uuid.UUID) ([]entities.BookCopy, error) {
	// Business Rule 1: Ensure the requested book title actually exists (returns 404 if not found)
	if _, err := s.repo.FindByID(ctx, bookID); err != nil {
		return nil, err
	}

	// Business Rule 2: Fetch all copies belonging to the validated book title
	return s.repo.FindCopiesByBookID(ctx, bookID)
}

// GetCopyByBarcode retrieves a physical copy by its unique barcode sticker.
func (s *bookService) GetCopyByBarcode(ctx context.Context, barcode string) (*entities.BookCopy, error) {
	// Direct lookup by barcode (repository handles eager loading of parent book details)
	return s.repo.FindCopyByBarcode(ctx, barcode)
}
