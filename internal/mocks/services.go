package mocks

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
)

type MockItemService struct {
	SearchFunc                   func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	GetByUniqueNameFunc          func(ctx context.Context, uniqueName string) (*models.Item, error)
	SearchReusableBlueprintsFunc func(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error)
}

func (m *MockItemService) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockItemService) GetByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	if m.GetByUniqueNameFunc != nil {
		return m.GetByUniqueNameFunc(ctx, uniqueName)
	}
	return nil, nil
}

func (m *MockItemService) SearchReusableBlueprints(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error) {
	if m.SearchReusableBlueprintsFunc != nil {
		return m.SearchReusableBlueprintsFunc(ctx, query, limit)
	}
	return nil, nil
}

type MockWishlistService struct {
	GetWishlistFunc    func(ctx context.Context, userID string) (*models.Wishlist, error)
	AddItemFunc        func(ctx context.Context, userID string, req models.AddItemRequest) error
	RemoveItemFunc     func(ctx context.Context, userID, uniqueName string) error
	UpdateQuantityFunc func(ctx context.Context, userID, uniqueName string, quantity int) error
}

func (m *MockWishlistService) GetWishlist(ctx context.Context, userID string) (*models.Wishlist, error) {
	if m.GetWishlistFunc != nil {
		return m.GetWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWishlistService) AddItem(ctx context.Context, userID string, req models.AddItemRequest) error {
	if m.AddItemFunc != nil {
		return m.AddItemFunc(ctx, userID, req)
	}
	return nil
}

func (m *MockWishlistService) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	if m.RemoveItemFunc != nil {
		return m.RemoveItemFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *MockWishlistService) UpdateQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	if m.UpdateQuantityFunc != nil {
		return m.UpdateQuantityFunc(ctx, userID, uniqueName, quantity)
	}
	return nil
}

type MockMaterialResolver struct {
	GetMaterialsFunc func(ctx context.Context, userID string) (*models.MaterialsResponse, error)
}

func (m *MockMaterialResolver) GetMaterials(ctx context.Context, userID string) (*models.MaterialsResponse, error) {
	if m.GetMaterialsFunc != nil {
		return m.GetMaterialsFunc(ctx, userID)
	}
	return nil, nil
}

type MockOwnedBlueprintsService struct {
	GetOwnedBlueprintsFunc func(ctx context.Context, userID string) (*models.OwnedBlueprints, error)
	AddBlueprintFunc       func(ctx context.Context, userID string, req models.AddBlueprintRequest) error
	RemoveBlueprintFunc    func(ctx context.Context, userID, uniqueName string) error
	BulkAddBlueprintsFunc  func(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error
	ClearAllBlueprintsFunc func(ctx context.Context, userID string) error
}

func (m *MockOwnedBlueprintsService) GetOwnedBlueprints(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
	if m.GetOwnedBlueprintsFunc != nil {
		return m.GetOwnedBlueprintsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockOwnedBlueprintsService) AddBlueprint(ctx context.Context, userID string, req models.AddBlueprintRequest) error {
	if m.AddBlueprintFunc != nil {
		return m.AddBlueprintFunc(ctx, userID, req)
	}
	return nil
}

func (m *MockOwnedBlueprintsService) RemoveBlueprint(ctx context.Context, userID, uniqueName string) error {
	if m.RemoveBlueprintFunc != nil {
		return m.RemoveBlueprintFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *MockOwnedBlueprintsService) BulkAddBlueprints(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error {
	if m.BulkAddBlueprintsFunc != nil {
		return m.BulkAddBlueprintsFunc(ctx, userID, req)
	}
	return nil
}

func (m *MockOwnedBlueprintsService) ClearAllBlueprints(ctx context.Context, userID string) error {
	if m.ClearAllBlueprintsFunc != nil {
		return m.ClearAllBlueprintsFunc(ctx, userID)
	}
	return nil
}
