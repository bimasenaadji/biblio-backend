package http

import (
	"biblio/internal/utils/response"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookHandler handles incoming HTTP requests for book resources.
type BookHandler struct {
	api *BookAPI
}

// NewBookHandler creates a new BookHandler instance.
func NewBookHandler(api *BookAPI) *BookHandler {
	return &BookHandler{
		api: api,
	}
}

// RegisterRoutes registers all book endpoint routes onto the Fiber router.
func (h *BookHandler) RegisterRoutes(router fiber.Router) {
	books := router.Group("/books")
	books.Post("", h.CreateBook)
	books.Get("", h.GetAllBooks)
	books.Get("/:id", h.GetBookByID)
	books.Put("/:id", h.UpdateBook)
	books.Delete("/:id", h.DeleteBook)
}

// CreateBook godoc
// @Summary      Create a new book
// @Description  Add a new book title catalog into the database
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        request body CreateBookRequest true "New book payload"
// @Success      201 {object} response.APIResponse{data=BookResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/books [post]
func (h *BookHandler) CreateBook(c *fiber.Ctx) error {
	var req CreateBookRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid JSON payload format", err.Error())
	}

	result, err := h.api.CreateBook(c.UserContext(), req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_BOOK_FAILED", "Failed to create book", err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "Book created successfully", result)
}

// GetAllBooks godoc
// @Summary      Get all books
// @Description  Retrieve all book catalog records from the database
// @Tags         Books
// @Produce      json
// @Success      200 {object} response.APIResponse{data=[]BookResponse}
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/books [get]
func (h *BookHandler) GetAllBooks(c *fiber.Ctx) error {
	result, err := h.api.GetAllBooks(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_BOOKS_FAILED", "Failed to fetch books", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Books fetched successfully", result)
}

// GetBookByID godoc
// @Summary      Get book by ID
// @Description  Retrieve detailed book information by its UUID
// @Tags         Books
// @Produce      json
// @Param        id path string true "Book UUID"
// @Success      200 {object} response.APIResponse{data=BookResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Router       /api/v1/books/{id} [get]
func (h *BookHandler) GetBookByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Book ID must be a valid UUID format", err.Error())
	}

	result, err := h.api.GetBookByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", "Book not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_BOOK_FAILED", "Failed to fetch book details", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Book details retrieved successfully", result)
}

// UpdateBook godoc
// @Summary      Update book
// @Description  Update book details by its UUID
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        id path string true "Book UUID"
// @Param        request body UpdateBookRequest true "Updated book payload"
// @Success      200 {object} response.APIResponse{data=BookResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Router       /api/v1/books/{id} [put]
func (h *BookHandler) UpdateBook(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Book ID must be a valid UUID format", err.Error())
	}

	var req UpdateBookRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid JSON payload format", err.Error())
	}

	result, err := h.api.UpdateBook(c.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", "Book not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_BOOK_FAILED", "Failed to update book", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Book updated successfully", result)
}

// DeleteBook godoc
// @Summary      Delete book by ID
// @Description  Soft delete a book catalog record by its UUID
// @Tags         Books
// @Produce      json
// @Param        id path string true "Book UUID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Router       /api/v1/books/{id} [delete]
func (h *BookHandler) DeleteBook(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Book ID must be a valid UUID format", err.Error())
	}

	if err := h.api.DeleteBook(c.UserContext(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", "Book not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_BOOK_FAILED", "Failed to delete book", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Book deleted successfully", nil)
}
