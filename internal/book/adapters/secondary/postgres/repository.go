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

// CreateCopy inserts a new physical book copy record into the database.
func (r *bookRepository) CreateCopy(ctx context.Context, copy *entities.BookCopy) error {

	// Directly persist the physical copy; BeforeCreate hook populates copy.ID
	return r.db.WithContext(ctx).Create(copy).Error
}

// FindCopiesByBookID retrieves all physical copies belonging to a specific book title.
func (r *bookRepository) FindCopiesByBookID(ctx context.Context, bookID uuid.UUID) ([]entities.BookCopy, error) {
	var copies []entities.BookCopy

	// Filter copies by foreign key book_id and sort chronologically
	err := r.db.WithContext(ctx).Where("book_id = ?", bookID).Order("created_at ASC").Find(&copies).Error
	if err != nil {
		return nil, err
	}
	return copies, nil
}

// FindCopyByBarcode retrieves a single book copy using its unique barcode sticker.
func (r *bookRepository) FindCopyByBarcode(ctx context.Context, barcode string) (*entities.BookCopy, error) {
	var copy entities.BookCopy

	// Preload("Book") eagerly joins the parent Book title record (Title, Author, ISBN)
	err := r.db.WithContext(ctx).Preload("Book").First(&copy, "barcode = ?", barcode).Error
	if err != nil {
		return nil, err
	}
	return &copy, nil
}

// UpdateCopyStatus updates only the circulation status of a book copy (e.g. available -> borrowed).
func (r *bookRepository) UpdateCopyStatus(ctx context.Context, id uuid.UUID, status entities.CopyStatus) error {

	// Execute a targeted UPDATE statement for the single status column without overwriting other fields
	return r.db.WithContext(ctx).Model(&entities.BookCopy{}).Where("id = ?", id).Update("status", status).Error
}
