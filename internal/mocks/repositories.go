package mocks

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
)

type MockItemRepository struct {
	SearchFunc           func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	FindByUniqueNameFunc func(ctx context.Context, uniqueName string) (*models.Item, error)
	FindByUniqueNamesFunc func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error)
}

func (m *MockItemRepository) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockItemRepository) FindByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	if m.FindByUniqueNameFunc != nil {
		return m.FindByUniqueNameFunc(ctx, uniqueName)
	}
	return nil, nil
}

func (m *MockItemRepository) FindByUniqueNames(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
	if m.FindByUniqueNamesFunc != nil {
		return m.FindByUniqueNamesFunc(ctx, uniqueNames)
	}
	return make(map[string]*models.Item), nil
}

type MockWishlistRepository struct {
	GetByUserIDFunc         func(ctx context.Context, userID string) (*models.Wishlist, error)
	CreateFunc              func(ctx context.Context, wishlist *models.Wishlist) error
	AddItemFunc             func(ctx context.Context, userID string, item models.WishlistItem) error
	RemoveItemFunc          func(ctx context.Context, userID, uniqueName string) error
	UpdateItemQuantityFunc  func(ctx context.Context, userID, uniqueName string, quantity int) error
	UpsertFunc              func(ctx context.Context, wishlist *models.Wishlist) error
}

func (m *MockWishlistRepository) GetByUserID(ctx context.Context, userID string) (*models.Wishlist, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWishlistRepository) Create(ctx context.Context, wishlist *models.Wishlist) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, wishlist)
	}
	return nil
}

func (m *MockWishlistRepository) AddItem(ctx context.Context, userID string, item models.WishlistItem) error {
	if m.AddItemFunc != nil {
		return m.AddItemFunc(ctx, userID, item)
	}
	return nil
}

func (m *MockWishlistRepository) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	if m.RemoveItemFunc != nil {
		return m.RemoveItemFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *MockWishlistRepository) UpdateItemQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	if m.UpdateItemQuantityFunc != nil {
		return m.UpdateItemQuantityFunc(ctx, userID, uniqueName, quantity)
	}
	return nil
}

func (m *MockWishlistRepository) Upsert(ctx context.Context, wishlist *models.Wishlist) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, wishlist)
	}
	return nil
}
