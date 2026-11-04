package response

import "github.com/gofiber/fiber/v2"

// PaginationMeta contains pagination details for list queries.
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
}

// ErrorDetail describes an error code and its optional metadata.
type ErrorDetail struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

// APIResponse is the unified enterprise JSON response envelope.
type APIResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    any             `json:"data,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
	Error   *ErrorDetail    `json:"error,omitempty"`
}

// Success returns a standardized 2xx success response.
func Success(c *fiber.Ctx, statusCode int, message string, data any) error {
	return c.Status(statusCode).JSON(APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Paginated returns a standardized 200 OK paginated response with metadata.
func Paginated(c *fiber.Ctx, message string, data any, meta PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

// Error returns a standardized error response with error code and details.
func Error(c *fiber.Ctx, statusCode int, code string, message string, details any) error {
	return c.Status(statusCode).JSON(APIResponse{
		Status:  "error",
		Message: message,
		Error: &ErrorDetail{
			Code:    code,
			Details: details,
		},
	})
}
