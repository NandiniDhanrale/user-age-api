package app

import (
	"context"
	"time"

	"github.com/NandiniDhanrale/user-age-api/config"
	"github.com/NandiniDhanrale/user-age-api/internal/handler"
	"github.com/NandiniDhanrale/user-age-api/internal/logger"
	"github.com/NandiniDhanrale/user-age-api/internal/repository"
	"github.com/NandiniDhanrale/user-age-api/internal/routes"
	"github.com/NandiniDhanrale/user-age-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	Fiber   *fiber.App
	Pool    *pgxpool.Pool
	DBReady bool
}

func New() *App {
	cfg := config.Load()
	log := logger.New()

	a := &App{}

	if cfg.DatabaseURL != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Warn("database unavailable, running without DB", zap.Error(err))
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = pool.Ping(ctx)
			cancel()
			if err != nil {
				log.Warn("database unreachable, running without DB", zap.Error(err))
			} else {
				a.Pool = pool
				a.DBReady = true
			}
		}
	} else {
		log.Warn("DATABASE_URL not set, running without DB")
	}

	var userSvc service.UserService
	if a.DBReady {
		userRepo := repository.NewUserRepository(a.Pool)
		userSvc = service.NewUserService(userRepo)
	} else {
		userSvc = service.NewUserService(nil)
	}

	userHandler := handler.NewUserHandler(userSvc)

	fb := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	fb.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "user-age-api",
			"version": "1.0.0",
			"health":  "/api/health",
			"docs":    "https://github.com/NandiniDhanrale/user-age-api",
		})
	})

	fb.Get("/api/health", func(c *fiber.Ctx) error {
		code := fiber.StatusOK
		status := "ok"
		if !a.DBReady {
			code = fiber.StatusServiceUnavailable
			status = "degraded"
		}
		return c.Status(code).JSON(fiber.Map{
			"status": status,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	routes.Setup(fb, userHandler, log)
	a.Fiber = fb

	return a
}
