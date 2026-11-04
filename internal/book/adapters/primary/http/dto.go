package http

import (
	"time"

	"github.com/google/uuid"
)

// 1. Book Title Request DTOs
// CreateBookRequest defines the incoming payload for creating a new book (POST /api/v1/books).
type CreateBookRequest struct {
	ISBN          string `json:"isbn" example:"978-0132350884"`
	Title         string `json:"title" example:"Clean Code"`
	Author        string `json:"author" example:"Robert C. Martin"`
	Category      string `json:"category" example:"Software Engineering"`
	PublishedYear int    `json:"published_year" example:"2008"`
}

// UpdateBookRequest defines the incoming payload for updating an existing book (PUT /api/v1/books/:id).
type UpdateBookRequest struct {
	ISBN          string `json:"isbn" example:"978-0132350884"`
	Title         string `json:"title" example:"Clean Code (Revised Edition)"`
	Author        string `json:"author" example:"Robert C. Martin"`
	Category      string `json:"category" example:"Software Engineering"`
	PublishedYear int    `json:"published_year" example:"2008"`
}

// BookResponse defines the outbound JSON payload delivered to clients.
type BookResponse struct {
	ID            uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ISBN          string    `json:"isbn" example:"978-0132350884"`
	Title         string    `json:"title" example:"Clean Code"`
	Author        string    `json:"author" example:"Robert C. Martin"`
	Category      string    `json:"category" example:"Software Engineering"`
	PublishedYear int       `json:"published_year" example:"2008"`
	CoverURL      string    `json:"cover_url" example:""`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 2. Physical Book Copy Request & Response DTOs
// CreateBookCopyRequest defines the incoming payload for adding a new physical copy to a book title.
// Note: ID, Status, and Timestamps are intentionally excluded to prevent mass assignment vulnerabilities.
type CreateBookCopyRequest struct {
	Barcode   string `json:"barcode" example:"BC-CLN-001"`
	RackCode  string `json:"rack_code" example:"RACK-A1"`
	Condition string `json:"condition" example:"good"`
}

// BookCopyResponse defines the outbound JSON payload for physical book copies.
// It optionally embeds parent Book details when eager loaded by the scanner endpoint.
type BookCopyResponse struct {
	ID        uuid.UUID     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BookID    uuid.UUID     `json:"book_id" example:"771e8400-e29b-41d4-a716-446655440001"`
	Barcode   string        `json:"barcode" example:"BC-CLN-001"`
	RackCode  string        `json:"rack_code" example:"RACK-A1"`
	Condition string        `json:"condition" example:"good"`
	Status    string        `json:"status" example:"available"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Book      *BookResponse `json:"book,omitempty"`
}
