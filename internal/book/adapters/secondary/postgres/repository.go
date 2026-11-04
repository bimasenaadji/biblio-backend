package postgres

import (
	"biblio/internal/book/core/entities"
	"biblio/internal/book/ports"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// bookRepository implements ports.BookRepository using GORM.
type bookRepository struct {
	db *gorm.DB
}

// NewBookRepository creates a new instance of the PostgreSQL book repository.
func NewBookRepository(db *gorm.DB) ports.BookRepository {
	return &bookRepository{
		db: db,
	}
}

// Create persists a new book entity into the database.
func (r *bookRepository) Create(ctx context.Context, book *entities.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

// FindAll retrieves all books ordered by creation timestamp descending.
func (r *bookRepository) FindAll(ctx context.Context) ([]entities.Book, error) {
	var books []entities.Book
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&books).Error
	if err != nil {
		return nil, err
	}
	return books, nil
}

// FindByID retrieves a single book by its unique UUID.
func (r *bookRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Book, error) {
	var book entities.Book
	err := r.db.WithContext(ctx).First(&book, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// Update saves modified fields of an existing book entity.
func (r *bookRepository) Update(ctx context.Context, book *entities.Book) error {
	return r.db.WithContext(ctx).Save(book).Error
}

// Delete soft-deletes a book record by UUID using GORM.
func (r *bookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Book{}, "id = ?", id).Error
}
