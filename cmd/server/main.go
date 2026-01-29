package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/graytonio/warframe-wishlist/internal/config"
	"github.com/graytonio/warframe-wishlist/internal/database"
	"github.com/graytonio/warframe-wishlist/internal/handlers"
	"github.com/graytonio/warframe-wishlist/internal/middleware"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/internal/services"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
)

func main() {
	cfg := config.Load()

	// Initialize logger with configured level (debug mode inferred from level)
	logger.Init(cfg.LogLevel)

	ctx := context.Background()
	logger.Info(ctx, "starting warframe-wishlist API server",
		"logLevel", cfg.LogLevel,
	)

	logger.Debug(ctx, "connecting to MongoDB", "uri", cfg.MongoURI, "database", cfg.MongoDatabase)
	db, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		logger.Error(ctx, "failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info(ctx, "connected to MongoDB")

	logger.Debug(ctx, "initializing repositories")
	itemRepo := repository.NewItemRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)

	logger.Debug(ctx, "initializing services")
	itemService := services.NewItemService(itemRepo)
	wishlistService := services.NewWishlistService(wishlistRepo, itemRepo)
	materialResolver := services.NewMaterialResolver(itemRepo, wishlistRepo)

	logger.Debug(ctx, "initializing handlers")
	healthHandler := handlers.NewHealthHandler()
	itemHandler := handlers.NewItemHandler(itemService)
	wishlistHandler := handlers.NewWishlistHandler(wishlistService, materialResolver)

	authMiddleware := middleware.NewAuthMiddleware(cfg.SupabaseJWTPublicKey)

	r := chi.NewRouter()

	// Middleware stack
	r.Use(chimiddleware.RequestID)      // Generate request IDs
	r.Use(middleware.LoggingMiddleware) // Custom structured logging
	r.Use(chimiddleware.Recoverer)      // Recover from panics

	allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/items", func(r chi.Router) {
			r.Get("/search", itemHandler.Search)
			r.Get("/*", itemHandler.GetByUniqueName)
		})

		r.Route("/wishlist", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/", wishlistHandler.GetWishlist)
			r.Post("/", wishlistHandler.AddItem)
			r.Get("/materials", wishlistHandler.GetMaterials)
			r.Delete("/*", wishlistHandler.RemoveItem)
			r.Patch("/*", wishlistHandler.UpdateQuantity)
		})
	})

	addr := ":" + cfg.ServerPort
	logger.Info(ctx, "server starting", "address", addr)

	// Graceful shutdown
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		logger.Info(ctx, "received shutdown signal", "signal", sig.String())
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error(ctx, "error during server shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(ctx, "server failed to start", "error", err)
		os.Exit(1)
	}

	logger.Info(ctx, "server stopped gracefully")
}
