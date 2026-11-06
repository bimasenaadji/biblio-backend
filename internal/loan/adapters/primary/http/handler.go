package http

import (
	"biblio/internal/utils/response"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoanHandler handles incoming HTTP requests for loan transactions.
type LoanHandler struct {
	api *LoanAPI
}

// NewLoanHandler creates a new instance of LoanHandler.
func NewLoanHandler(api *LoanAPI) *LoanHandler {
	return &LoanHandler{
		api: api,
	}
}

// RegisterRoutes registers loan endpoints onto the provided Fiber router group.
func (h *LoanHandler) RegisterRoutes(router fiber.Router) {
	loans := router.Group("/loans")

	loans.Post("", h.BorrowBooks)
	loans.Get("/:id", h.GetLoanByID)
	loans.Get("/number/:loanNumber", h.GetLoanByNumber)
	loans.Post("/:id/return", h.ReturnBooks)
}

// BorrowBooks godoc
// @Summary      Borrow books
// @Description  Create a new loan transaction by scanning up to 5 physical book copy barcodes
// @Tags         Loans
// @Accept       json
// @Produce      json
// @Param        request body BorrowBookRequest true "Borrow books payload"
// @Success      201 {object} response.APIResponse{data=LoanResponse}
// @Failure      400 {object} response.APIResponse
// @Failure      500 {object} response.APIResponse
// @Router       /api/v1/loans [post]
func (h *LoanHandler) BorrowBooks(c *fiber.Ctx) error {
	var req BorrowBookRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid request payload format", err.Error())
	}

	res, err := h.api.BorrowBooks(c.UserContext(), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BORROW_FAILED", "Failed to process book loan", err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "Loan transaction completed successfully", res)
}

// GetLoanByID godoc
// @Summary      Get loan by ID
// @Description  Retrieve full loan transaction receipt and its borrowed items by UUID
// @Tags         Loans
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Loan UUID"
// @Success      200  {object}  response.APIResponse{data=LoanResponse}
// @Failure      400  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Failure      500  {object}  response.APIResponse
// @Router       /api/v1/loans/{id} [get]
func (h *LoanHandler) GetLoanByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid loan UUID format", err.Error())
	}

	res, err := h.api.GetLoanByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Loan record not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve loan record", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Loan retrieved successfully", res)
}

// GetLoanByNumber godoc
// @Summary      Get loan by loan number
// @Description  Retrieve full loan transaction receipt by reference number (e.g. LN-YYYYMMDD-XXXX)
// @Tags         Loans
// @Accept       json
// @Produce      json
// @Param        loanNumber   path      string  true  "Loan Reference Number"
// @Success      200          {object}  response.APIResponse{data=LoanResponse}
// @Failure      400          {object}  response.APIResponse
// @Failure      404          {object}  response.APIResponse
// @Failure      500          {object}  response.APIResponse
// @Router       /api/v1/loans/number/{loanNumber} [get]
func (h *LoanHandler) GetLoanByNumber(c *fiber.Ctx) error {
	loanNumber := c.Params("loanNumber")
	if loanNumber == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PARAM", "Loan number parameter is required", nil)
	}

	res, err := h.api.GetLoanByNumber(c.UserContext(), loanNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Loan record not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve loan record", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Loan retrieved successfully", res)
}

// ReturnBooks godoc
// @Summary      Return borrowed books
// @Description  Process book returns for an active loan by specifying returned barcodes and physical condition
// @Tags         Loans
// @Accept       json
// @Produce      json
// @Param        id      path      string              true  "Loan UUID"
// @Param        request body      ReturnBooksRequest  true  "Return books payload"
// @Success      200     {object}  response.APIResponse{data=LoanResponse}
// @Failure      400     {object}  response.APIResponse
// @Failure      404     {object}  response.APIResponse
// @Failure      500     {object}  response.APIResponse
// @Router       /api/v1/loans/{id}/return [post]
func (h *LoanHandler) ReturnBooks(c *fiber.Ctx) error {
	idParam := c.Params("id")
	loanID, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid loan UUID format", err.Error())
	}

	var req ReturnBooksRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "Invalid request payload format", err.Error())
	}

	res, err := h.api.ReturnBooks(c.UserContext(), loanID, req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "RETURN_FAILED", "Failed to process book return", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Books returned successfully", res)
}
