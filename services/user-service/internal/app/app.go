package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yakonstantine/go-msa-lab/services/user-service/config"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/handler"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/handler/middleware"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/infra/repo"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase/user"
)

func Run(cfg *config.Config) {
	txf := &repo.TransactionFactoryMemo{}
	ur := repo.NewUserMemoRepo()
	sr := repo.NewSMTPMemoRepo()

	userUseCase := user.NewUseCase(txf, ur, sr)
	userHandler := handler.NewUserHandler(userUseCase)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorMiddleware())

	g := router.Group("/api")
	g.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	handler.NewUserRoutes(g, userHandler)

	runServer(router, "8011")
}

func runServer(router *gin.Engine, port string) {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until OS signal
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}

	slog.Info("server exited gracefully")
}
