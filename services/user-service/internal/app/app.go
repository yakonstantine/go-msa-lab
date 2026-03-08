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
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/infra/repo"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase/user"
)

func Run(cfg *config.Config) {
	txf := &repo.TransactionFactoryMemo{}
	ur := repo.NewUserMemoRepo()
	sr := repo.NewSMTPMemoRepo()

	userUseCase := user.New(txf, ur, sr)
	userHandler := handler.NewUserHandler(userUseCase)

	router := gin.Default()

	g := router.Group("/api")
	g.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	handler.NewUserRoutes(g, userHandler)

	runServer(router)
}

func runServer(router *gin.Engine) {
	srv := &http.Server{
		Addr:    ":8011",
		Handler: router,
	}

	go func() {
		slog.Info("starting server on :8011")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until OS signal
	quit := make(chan os.Signal, 1)
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
