package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/mocks"
	"github.com/graytonio/warframe-wishlist/internal/models"
)

func TestMaterialResolver_GetMaterials_EmptyWishlist(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return nil, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if len(result.Materials) != 0 {
		t.Errorf("expected empty materials, got %d", len(result.Materials))
	}

	if result.TotalCredits != 0 {
		t.Errorf("expected 0 credits, got %d", result.TotalCredits)
	}
}

func TestMaterialResolver_GetMaterials_WishlistWithEmptyItems(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items:  []models.WishlistItem{},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Materials) != 0 {
		t.Errorf("expected empty materials, got %d", len(result.Materials))
	}
}

func TestMaterialResolver_GetMaterials_SimpleItem(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{
				"/Lotus/Item1": {
					UniqueName: "/Lotus/Item1",
					Name:       "Simple Item",
					BuildPrice: 1000,
					Components: []models.Component{},
				},
			}, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Materials) != 1 {
		t.Errorf("expected 1 material, got %d", len(result.Materials))
	}

	if result.TotalCredits != 1000 {
		t.Errorf("expected 1000 credits, got %d", result.TotalCredits)
	}
}

func TestMaterialResolver_GetMaterials_ItemWithComponents(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{
				"/Lotus/Warframe": {
					UniqueName: "/Lotus/Warframe",
					Name:       "Test Warframe",
					BuildPrice: 25000,
					Components: []models.Component{
						{UniqueName: "/Lotus/Resource1", Name: "Resource 1", ItemCount: 100},
						{UniqueName: "/Lotus/Resource2", Name: "Resource 2", ItemCount: 50},
					},
				},
			}, nil
		},
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			return nil, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Warframe", Quantity: 1, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Materials) != 2 {
		t.Errorf("expected 2 materials, got %d", len(result.Materials))
	}

	if result.TotalCredits != 25000 {
		t.Errorf("expected 25000 credits, got %d", result.TotalCredits)
	}

	materialCounts := make(map[string]int)
	for _, mat := range result.Materials {
		materialCounts[mat.UniqueName] = mat.TotalCount
	}

	if materialCounts["/Lotus/Resource1"] != 100 {
		t.Errorf("expected 100 Resource1, got %d", materialCounts["/Lotus/Resource1"])
	}
	if materialCounts["/Lotus/Resource2"] != 50 {
		t.Errorf("expected 50 Resource2, got %d", materialCounts["/Lotus/Resource2"])
	}
}

func TestMaterialResolver_GetMaterials_MultipleQuantity(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{
				"/Lotus/Item1": {
					UniqueName: "/Lotus/Item1",
					Name:       "Simple Item",
					BuildPrice: 1000,
					Components: []models.Component{
						{UniqueName: "/Lotus/Resource1", Name: "Resource 1", ItemCount: 10},
					},
				},
			}, nil
		},
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			return nil, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 3, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCredits != 3000 {
		t.Errorf("expected 3000 credits (1000 * 3), got %d", result.TotalCredits)
	}

	for _, mat := range result.Materials {
		if mat.UniqueName == "/Lotus/Resource1" && mat.TotalCount != 30 {
			t.Errorf("expected 30 Resource1 (10 * 3), got %d", mat.TotalCount)
		}
	}
}

func TestMaterialResolver_GetMaterials_NestedComponents(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{
				"/Lotus/Warframe": {
					UniqueName: "/Lotus/Warframe",
					Name:       "Test Warframe",
					BuildPrice: 25000,
					Components: []models.Component{
						{UniqueName: "/Lotus/Chassis", Name: "Chassis", ItemCount: 1},
					},
				},
			}, nil
		},
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			if uniqueName == "/Lotus/Chassis" {
				return &models.Item{
					UniqueName: "/Lotus/Chassis",
					Name:       "Chassis",
					BuildPrice: 15000,
					Components: []models.Component{
						{UniqueName: "/Lotus/Alloy", Name: "Alloy Plate", ItemCount: 500},
					},
				}, nil
			}
			return nil, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Warframe", Quantity: 1, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCredits != 40000 {
		t.Errorf("expected 40000 credits (25000 + 15000), got %d", result.TotalCredits)
	}

	for _, mat := range result.Materials {
		if mat.UniqueName == "/Lotus/Alloy" && mat.TotalCount != 500 {
			t.Errorf("expected 500 Alloy Plate, got %d", mat.TotalCount)
		}
	}
}

func TestMaterialResolver_GetMaterials_RepositoryError(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return nil, errors.New("database error")
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	_, err := resolver.GetMaterials(context.Background(), "user-123")

	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestMaterialResolver_GetMaterials_ItemNotInRepository(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{}, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/NonExistent", Quantity: 1, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Materials) != 0 {
		t.Errorf("expected 0 materials for non-existent item, got %d", len(result.Materials))
	}
}

func TestMaterialResolver_GetMaterials_CycleDetection(t *testing.T) {
	callCount := 0
	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
			return map[string]*models.Item{
				"/Lotus/ItemA": {
					UniqueName: "/Lotus/ItemA",
					Name:       "Item A",
					BuildPrice: 1000,
					Components: []models.Component{
						{UniqueName: "/Lotus/ItemB", Name: "Item B", ItemCount: 1},
					},
				},
			}, nil
		},
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			callCount++
			if callCount > 100 {
				t.Fatal("too many recursive calls, cycle detection may have failed")
			}
			if uniqueName == "/Lotus/ItemB" {
				return &models.Item{
					UniqueName: "/Lotus/ItemB",
					Name:       "Item B",
					BuildPrice: 500,
					Components: []models.Component{
						{UniqueName: "/Lotus/ItemA", Name: "Item A", ItemCount: 1},
					},
				}, nil
			}
			return nil, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{
				UserID: userID,
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/ItemA", Quantity: 1, AddedAt: time.Now()},
				},
			}, nil
		},
	}

	resolver := NewMaterialResolver(mockItemRepo, mockWishlistRepo)
	result, err := resolver.GetMaterials(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}
}
