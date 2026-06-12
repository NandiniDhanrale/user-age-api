package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NandiniDhanrale/user-age-api/config"
	"github.com/NandiniDhanrale/user-age-api/app"
	"github.com/NandiniDhanrale/user-age-api/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New()
	a := app.New()

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort)
	log.Info("starting server", zap.String("addr", addr))

	go func() {
		if err := a.Fiber.Listen(addr); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Fiber.ShutdownWithContext(ctx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
	if a.Pool != nil {
		a.Pool.Close()
	}
}
