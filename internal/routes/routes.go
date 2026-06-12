package routes

import (
	"github.com/NandiniDhanrale/user-age-api/internal/handler"
	"github.com/NandiniDhanrale/user-age-api/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Setup(app *fiber.App, userHandler *handler.UserHandler, log *zap.Logger) {
	api := app.Group("/api", middleware.RequestID(), middleware.RequestLogger(log), middleware.GlobalErrorHandler(log))

	users := api.Group("/users")

	users.Post("/", userHandler.Create)
	users.Get("/", userHandler.List)
	users.Get("/:id", userHandler.GetByID)
	users.Put("/:id", userHandler.Update)
	users.Delete("/:id", userHandler.Delete)
}
