package services

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
)

type ItemServiceInterface interface {
	Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	GetByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error)
}

type WishlistServiceInterface interface {
	GetWishlist(ctx context.Context, userID string) (*models.Wishlist, error)
	AddItem(ctx context.Context, userID string, req models.AddItemRequest) error
	RemoveItem(ctx context.Context, userID, uniqueName string) error
	UpdateQuantity(ctx context.Context, userID, uniqueName string, quantity int) error
}

type MaterialResolverInterface interface {
	GetMaterials(ctx context.Context, userID string) (*models.MaterialsResponse, error)
}

var _ ItemServiceInterface = (*ItemService)(nil)
var _ WishlistServiceInterface = (*WishlistService)(nil)
var _ MaterialResolverInterface = (*MaterialResolver)(nil)
