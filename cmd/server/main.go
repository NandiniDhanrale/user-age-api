package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func main() {
	cfg := config.Load()
	log := logger.New()

	var userSvc service.UserService
	dbAvailable := false

	if cfg.DatabaseURL != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Warn("database unavailable, running without DB", zap.Error(err))
		} else {
			pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
			if err != nil {
				log.Warn("database unreachable, running without DB", zap.Error(err))
			} else {
				dbAvailable = true
				userRepo := repository.NewUserRepository(pool)
				userSvc = service.NewUserService(userRepo)
				defer pool.Close()
			}
		}
	} else {
		log.Warn("DATABASE_URL not set, running without DB")
	}

	if userSvc == nil {
		userSvc = service.NewUserService(nil)
	}

	userHandler := handler.NewUserHandler(userSvc)

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		status := "ok"
		code := http.StatusOK
		if !dbAvailable {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		return c.Status(code).JSON(fiber.Map{
			"status": status,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	routes.Setup(app, userHandler, log)

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort)
	log.Info("starting server", zap.String("addr", addr))

	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
}
