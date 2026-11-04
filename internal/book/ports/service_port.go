package ports

import (
	"biblio/internal/book/core/entities"
	"context"

	"github.com/google/uuid"
)

// BookService defines the primary (inbound) API port for book business use cases.
type BookService interface {
	CreateBook(ctx context.Context, book *entities.Book) (*entities.Book, error)
	GetAllBooks(ctx context.Context) ([]entities.Book, error)
	GetBookByID(ctx context.Context, id uuid.UUID) (*entities.Book, error)
	UpdateBook(ctx context.Context, id uuid.UUID, input *entities.Book) (*entities.Book, error)
	DeleteBook(ctx context.Context, id uuid.UUID) error
}
