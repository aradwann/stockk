package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/hibiken/asynq"
	slogchi "github.com/samber/slog-chi"

	"stockk/internal/config"
	"stockk/internal/controllers"
	"stockk/internal/db"
	"stockk/internal/repository"
	"stockk/internal/service"
	"stockk/internal/worker"
)

func main() {

	// Initialize configuration
	cfg, err := config.LoadConfig(".", ".env")
	if err != nil {
		slog.Error("Server failed", "error", err)
	}

	// Create logger
	logger := config.CreateLogger(cfg.Environment)
	slog.SetDefault(logger)

	// Initialize Database
	dbConn := db.InitDatabase(cfg)
	defer dbConn.Close()

	// Set up Redis options for task distribution.
	RedisClientOpts := asynq.RedisClientOpt{Addr: cfg.RedisAddress}
	asynqClient := asynq.NewClient(RedisClientOpts)
	defer asynqClient.Close()

	// Initialize repositories
	ingredientRepo := repository.NewIngredientRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	productRepo := repository.NewProductRepository(dbConn)
	taskQueueRepo := repository.NewTaskQueueRepository(asynqClient)
	// Initialize services
	orderService := service.NewOrderService(orderRepo, productRepo, ingredientRepo)
	ingredientService := service.NewIngredientService(ingredientRepo, taskQueueRepo)

	// Initialize controllers
	orderController := controllers.NewOrderController(orderService, ingredientService)

	go worker.RunTaskProcessor(cfg, RedisClientOpts, ingredientRepo)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(slogchi.New(logger))
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Add timeout
	r.Use(middleware.Timeout(60 * time.Second))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/orders", orderController.CreateOrder)

	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Server configuration
	serverAddr := cfg.HTTPServerAddress

	// Create server
	server := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting server", "address", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	logger.Info("Server exited")
}
