package ports

import (
	"biblio/internal/book/core/entities"
	"context"

	"github.com/google/uuid"
)

// BookRepository defines the secondary (outbound) SPI port for data persistence.
type BookRepository interface {
	Create(ctx context.Context, book *entities.Book) error
	FindAll(ctx context.Context) ([]entities.Book, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Book, error)
	Update(ctx context.Context, book *entities.Book) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Book copy operations
	CreateCopy(ctx context.Context, copy *entities.BookCopy) error
	FindCopiesByBookID(ctx context.Context, bookID uuid.UUID) ([]entities.BookCopy, error)
	FindCopyByBarcode(ctx context.Context, barcode string) (*entities.BookCopy, error)
	UpdateCopyStatus(ctx context.Context, id uuid.UUID, status entities.CopyStatus) error
}
