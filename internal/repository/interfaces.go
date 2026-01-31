package repository

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
)

type ItemRepositoryInterface interface {
	Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	FindByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error)
	FindByUniqueNames(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error)
	SearchReusableBlueprints(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error)
}

type WishlistRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID string) (*models.Wishlist, error)
	Create(ctx context.Context, wishlist *models.Wishlist) error
	AddItem(ctx context.Context, userID string, item models.WishlistItem) error
	RemoveItem(ctx context.Context, userID, uniqueName string) error
	UpdateItemQuantity(ctx context.Context, userID, uniqueName string, quantity int) error
	Upsert(ctx context.Context, wishlist *models.Wishlist) error
}

type OwnedBlueprintsRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID string) (*models.OwnedBlueprints, error)
	Create(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error
	AddBlueprint(ctx context.Context, userID string, blueprint models.OwnedBlueprint) error
	RemoveBlueprint(ctx context.Context, userID, uniqueName string) error
	BulkAddBlueprints(ctx context.Context, userID string, blueprints []models.OwnedBlueprint) error
	ClearAll(ctx context.Context, userID string) error
}

var _ ItemRepositoryInterface = (*ItemRepository)(nil)
var _ WishlistRepositoryInterface = (*WishlistRepository)(nil)
var _ OwnedBlueprintsRepositoryInterface = (*OwnedBlueprintsRepository)(nil)
