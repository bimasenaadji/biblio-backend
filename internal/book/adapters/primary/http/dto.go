package http

import (
	"time"

	"github.com/google/uuid"
)

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
