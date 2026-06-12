package middleware

import (
	"errors"

	"github.com/NandiniDhanrale/user-age-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func GlobalErrorHandler(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err == nil {
			return nil
		}

		rid, _ := c.Locals("request_id").(string)
		log.Error("request error",
			zap.String("request_id", rid),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Error(err),
		)

		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
		}

		if errors.Is(err, service.ErrInvalidDOB) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_dob",
				Message: err.Error(),
			})
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(ErrorResponse{
				Error:   "fiber_error",
				Message: fiberErr.Message,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "something went wrong",
		})
	}
}
