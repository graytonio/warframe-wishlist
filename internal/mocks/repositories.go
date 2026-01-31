package mocks

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
)

type MockItemRepository struct {
	SearchFunc                   func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	FindByUniqueNameFunc         func(ctx context.Context, uniqueName string) (*models.Item, error)
	FindByUniqueNamesFunc        func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error)
	SearchReusableBlueprintsFunc func(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error)
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

func (m *MockItemRepository) SearchReusableBlueprints(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error) {
	if m.SearchReusableBlueprintsFunc != nil {
		return m.SearchReusableBlueprintsFunc(ctx, query, limit)
	}
	return nil, nil
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

type MockOwnedBlueprintsRepository struct {
	GetByUserIDFunc       func(ctx context.Context, userID string) (*models.OwnedBlueprints, error)
	CreateFunc            func(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error
	AddBlueprintFunc      func(ctx context.Context, userID string, blueprint models.OwnedBlueprint) error
	RemoveBlueprintFunc   func(ctx context.Context, userID, uniqueName string) error
	BulkAddBlueprintsFunc func(ctx context.Context, userID string, blueprints []models.OwnedBlueprint) error
	ClearAllFunc          func(ctx context.Context, userID string) error
}

func (m *MockOwnedBlueprintsRepository) GetByUserID(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockOwnedBlueprintsRepository) Create(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ownedBlueprints)
	}
	return nil
}

func (m *MockOwnedBlueprintsRepository) AddBlueprint(ctx context.Context, userID string, blueprint models.OwnedBlueprint) error {
	if m.AddBlueprintFunc != nil {
		return m.AddBlueprintFunc(ctx, userID, blueprint)
	}
	return nil
}

func (m *MockOwnedBlueprintsRepository) RemoveBlueprint(ctx context.Context, userID, uniqueName string) error {
	if m.RemoveBlueprintFunc != nil {
		return m.RemoveBlueprintFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *MockOwnedBlueprintsRepository) BulkAddBlueprints(ctx context.Context, userID string, blueprints []models.OwnedBlueprint) error {
	if m.BulkAddBlueprintsFunc != nil {
		return m.BulkAddBlueprintsFunc(ctx, userID, blueprints)
	}
	return nil
}

func (m *MockOwnedBlueprintsRepository) ClearAll(ctx context.Context, userID string) error {
	if m.ClearAllFunc != nil {
		return m.ClearAllFunc(ctx, userID)
	}
	return nil
}
