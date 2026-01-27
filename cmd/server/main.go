package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/graytonio/warframe-wishlist/internal/config"
	"github.com/graytonio/warframe-wishlist/internal/database"
	"github.com/graytonio/warframe-wishlist/internal/handlers"
	"github.com/graytonio/warframe-wishlist/internal/middleware"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	log.Println("Connected to MongoDB")

	itemRepo := repository.NewItemRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)

	itemService := services.NewItemService(itemRepo)
	wishlistService := services.NewWishlistService(wishlistRepo, itemRepo)
	materialResolver := services.NewMaterialResolver(itemRepo, wishlistRepo)

	healthHandler := handlers.NewHealthHandler()
	itemHandler := handlers.NewItemHandler(itemService)
	wishlistHandler := handlers.NewWishlistHandler(wishlistService, materialResolver)

	authMiddleware := middleware.NewAuthMiddleware(cfg.SupabaseJWTSecret)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

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
			r.Get("/{uniqueName}", itemHandler.GetByUniqueName)
		})

		r.Route("/wishlist", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/", wishlistHandler.GetWishlist)
			r.Post("/", wishlistHandler.AddItem)
			r.Delete("/{uniqueName}", wishlistHandler.RemoveItem)
			r.Patch("/{uniqueName}", wishlistHandler.UpdateQuantity)
			r.Get("/materials", wishlistHandler.GetMaterials)
		})
	})

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
