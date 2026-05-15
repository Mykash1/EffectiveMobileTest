package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"EffectiveMobileTest/internal/config"
	"EffectiveMobileTest/internal/handler"
	"EffectiveMobileTest/internal/logger"
	"EffectiveMobileTest/internal/middleware"
	"EffectiveMobileTest/internal/repository"
	"EffectiveMobileTest/internal/service"

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

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	appLogger.Info("database connected")

	repo := repository.NewSubscriptionRepository(db)
	serviceLayer := service.NewSubscriptionService(repo, appLogger)
	handlerLayer := handler.NewSubscriptionHandler(serviceLayer, appLogger)

	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.LoggingMiddleware(appLogger))

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", handlerLayer.Create)
		r.Get("/", handlerLayer.List)
		r.Get("/total", handlerLayer.Total)
		r.Get("/{id}", handlerLayer.GetByID)
		r.Put("/{id}", handlerLayer.Update)
		r.Delete("/{id}", handlerLayer.Delete)
	})

	r.Handle("/swagger/*", http.StripPrefix("/swagger/", handler.SwaggerUI()))
	r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("server started", "port", cfg.AppPort)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	appLogger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("server forced to shutdown", "error", err)
	}

	appLogger.Info("server stopped")
}
