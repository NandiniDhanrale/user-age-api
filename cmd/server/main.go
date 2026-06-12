package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("failed to ping database", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	routes.Setup(app, userHandler, log)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.ServerPort)
		log.Info("starting server", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	if err := app.Shutdown(); err != nil {
		log.Fatal("shutdown error", zap.Error(err))
	}
}
