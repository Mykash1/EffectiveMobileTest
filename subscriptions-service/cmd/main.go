package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"EffectiveMobileTest/subscriptions-service/internal/config"
	"EffectiveMobileTest/subscriptions-service/internal/handler"
	"EffectiveMobileTest/subscriptions-service/internal/logger"
	"EffectiveMobileTest/subscriptions-service/internal/middleware"
	"EffectiveMobileTest/subscriptions-service/internal/repository"
	"EffectiveMobileTest/subscriptions-service/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.New()

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewSubscriptionRepository(db)
	serviceLayer := service.NewSubscriptionService(repo)
	handlerLayer := handler.NewSubscriptionHandler(serviceLayer)

	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.LoggingMiddleware(appLogger))

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", handlerLayer.Create)
		r.Get("/", handlerLayer.List)
		r.Get("/total", handlerLayer.Total)
		r.Get("/{id}", handlerLayer.GetByID)
		r.Delete("/{id}", handlerLayer.Delete)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		appLogger.Info("server started", "port", cfg.AppPort)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)

	appLogger.Info("server stopped")
}
