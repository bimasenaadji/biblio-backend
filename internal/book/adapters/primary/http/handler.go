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
	// Nested routes under /api/v1/books
	books := router.Group("/books")
	books.Post("", h.CreateBook)
	books.Get("", h.GetAllBooks)
	books.Get("/:id", h.GetBookByID)
	books.Put("/:id", h.UpdateBook)
	books.Delete("/:id", h.DeleteBook)

	// Physical Book Copies nested endpoints under /api/v1/books/:id/copies
	books.Post("/:id/copies", h.AddBookCopy)
	books.Get("/:id/copies", h.GetCopiesByBookID)

	// Direct barcode scanner lookup endpoint attached directly to /api/v1/copies/barcode/:barcode
	router.Get("/copies/barcode/:barcode", h.GetCopyByBarcode)
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

// AddBookCopy godoc
// @Summary      Add physical copy to a book
// @Description  Register a new physical book copy with a unique barcode and rack location
// @Tags         Book Copies
// @Accept       json
// @Produce      json
// @Param        id path string true "Book UUID"
// @Param        request body CreateBookCopyRequest true "Physical copy payload"
// @Success      201 {object} response.APIResponse{data=BookCopyResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/books/{id}/copies [post]
func (h *BookHandler) AddBookCopy(c *fiber.Ctx) error {
	// Step 1: Parse and validate parent book UUID from URL route param
	bookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Book ID must be a valid UUID format", err.Error())
	}

	// Step 2: Parse incoming JSON request body into DTO
	var req CreateBookCopyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid JSON payload format", err.Error())
	}

	// Step 3: Forward request with standard context to API coordinator
	result, err := h.api.AddBookCopy(c.UserContext(), bookID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", "Parent book title not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_COPY_FAILED", "Failed to add book copy", err.Error())
	}

	// Step 4: Return 201 Created with standard success envelope
	return response.Success(c, fiber.StatusCreated, "Book copy added successfully", result)
}

// GetCopiesByBookID godoc
// @Summary      Get all physical copies of a book
// @Description  Retrieve all physical item copies registered under a specific book title
// @Tags         Book Copies
// @Produce      json
// @Param        id path string true "Book UUID"
// @Success      200 {object} response.APIResponse{data=[]BookCopyResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/books/{id}/copies [get]
func (h *BookHandler) GetCopiesByBookID(c *fiber.Ctx) error {
	// Step 1: Parse and validate book UUID from URL route param
	bookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Book ID must be a valid UUID format", err.Error())
	}

	// Step 2: Delegate retrieval to API coordinator
	result, err := h.api.GetCopiesByBookID(c.UserContext(), bookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", "Parent book title not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_COPIES_FAILED", "Failed to fetch book copies", err.Error())
	}

	// Step 3: Return 200 OK with list of copies
	return response.Success(c, fiber.StatusOK, "Book copies retrieved successfully", result)
}

// GetCopyByBarcode godoc
// @Summary      Get book copy by barcode (Scan)
// @Description  Lookup physical book copy and its parent book details by scanning unique barcode
// @Tags         Book Copies
// @Produce      json
// @Param        barcode path string true "Unique Barcode (e.g. BC-CLN-001)"
// @Success      200 {object} response.APIResponse{data=BookCopyResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      404 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/copies/barcode/{barcode} [get]
func (h *BookHandler) GetCopyByBarcode(c *fiber.Ctx) error {
	// Step 1: Extract barcode string from URL route param
	barcode := c.Params("barcode")
	if barcode == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BARCODE", "Barcode parameter is required", nil)
	}

	// Step 2: Query copy by barcode through API coordinator
	result, err := h.api.GetCopyByBarcode(c.UserContext(), barcode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "COPY_NOT_FOUND", "Book copy with specified barcode not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_COPY_FAILED", "Failed to lookup barcode", err.Error())
	}

	// Step 3: Return 200 OK with copy and preloaded book title data
	return response.Success(c, fiber.StatusOK, "Book copy found successfully", result)
}
