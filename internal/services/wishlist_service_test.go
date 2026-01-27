package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/mocks"
	"github.com/graytonio/warframe-wishlist/internal/models"
)

func TestWishlistService_GetWishlist(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *models.Wishlist
		mockError      error
		expectError    bool
		expectNewEmpty bool
	}{
		{
			name:   "existing wishlist found",
			userID: "user-123",
			mockReturn: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			mockError:      nil,
			expectError:    false,
			expectNewEmpty: false,
		},
		{
			name:           "no wishlist returns empty",
			userID:         "new-user",
			mockReturn:     nil,
			mockError:      nil,
			expectError:    false,
			expectNewEmpty: true,
		},
		{
			name:        "repository error",
			userID:      "error-user",
			mockReturn:  nil,
			mockError:   errors.New("database error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishlistRepo := &mocks.MockWishlistRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
					return tt.mockReturn, tt.mockError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewWishlistService(mockWishlistRepo, mockItemRepo)
			wishlist, err := service.GetWishlist(context.Background(), tt.userID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && wishlist == nil {
				t.Error("expected wishlist but got nil")
			}
			if tt.expectNewEmpty && wishlist != nil && len(wishlist.Items) != 0 {
				t.Errorf("expected empty items, got %d", len(wishlist.Items))
			}
		})
	}
}

func TestWishlistService_AddItem(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		request          models.AddItemRequest
		mockItem         *models.Item
		mockWishlist     *models.Wishlist
		itemError        error
		wishlistError    error
		createError      error
		addItemError     error
		expectError      error
	}{
		{
			name:   "add item to new wishlist",
			userID: "user-123",
			request: models.AddItemRequest{
				UniqueName: "/Lotus/Item1",
				Quantity:   1,
			},
			mockItem:     &models.Item{UniqueName: "/Lotus/Item1", Name: "Item 1"},
			mockWishlist: nil,
			expectError:  nil,
		},
		{
			name:   "add item to existing wishlist",
			userID: "user-123",
			request: models.AddItemRequest{
				UniqueName: "/Lotus/Item2",
				Quantity:   2,
			},
			mockItem: &models.Item{UniqueName: "/Lotus/Item2", Name: "Item 2"},
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: nil,
		},
		{
			name:   "item not found error",
			userID: "user-123",
			request: models.AddItemRequest{
				UniqueName: "/Lotus/Nonexistent",
			},
			mockItem:    nil,
			expectError: ErrItemNotFound,
		},
		{
			name:   "item already in wishlist",
			userID: "user-123",
			request: models.AddItemRequest{
				UniqueName: "/Lotus/Item1",
			},
			mockItem: &models.Item{UniqueName: "/Lotus/Item1", Name: "Item 1"},
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: ErrItemAlreadyInWishlist,
		},
		{
			name:   "default quantity when zero",
			userID: "user-123",
			request: models.AddItemRequest{
				UniqueName: "/Lotus/Item1",
				Quantity:   0,
			},
			mockItem:     &models.Item{UniqueName: "/Lotus/Item1", Name: "Item 1"},
			mockWishlist: nil,
			expectError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockItemRepo := &mocks.MockItemRepository{
				FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
					return tt.mockItem, tt.itemError
				},
			}
			mockWishlistRepo := &mocks.MockWishlistRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
					return tt.mockWishlist, tt.wishlistError
				},
				CreateFunc: func(ctx context.Context, wishlist *models.Wishlist) error {
					return tt.createError
				},
				AddItemFunc: func(ctx context.Context, userID string, item models.WishlistItem) error {
					return tt.addItemError
				},
			}

			service := NewWishlistService(mockWishlistRepo, mockItemRepo)
			err := service.AddItem(context.Background(), tt.userID, tt.request)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectError)
				} else if !errors.Is(err, tt.expectError) {
					t.Errorf("expected error %v but got %v", tt.expectError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWishlistService_RemoveItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		uniqueName    string
		mockWishlist  *models.Wishlist
		wishlistError error
		removeError   error
		expectError   error
	}{
		{
			name:       "successfully remove item",
			userID:     "user-123",
			uniqueName: "/Lotus/Item1",
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: nil,
		},
		{
			name:         "no wishlist exists",
			userID:       "user-123",
			uniqueName:   "/Lotus/Item1",
			mockWishlist: nil,
			expectError:  ErrItemNotInWishlist,
		},
		{
			name:       "item not in wishlist",
			userID:     "user-123",
			uniqueName: "/Lotus/Item2",
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: ErrItemNotInWishlist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishlistRepo := &mocks.MockWishlistRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
					return tt.mockWishlist, tt.wishlistError
				},
				RemoveItemFunc: func(ctx context.Context, userID, uniqueName string) error {
					return tt.removeError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewWishlistService(mockWishlistRepo, mockItemRepo)
			err := service.RemoveItem(context.Background(), tt.userID, tt.uniqueName)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectError)
				} else if !errors.Is(err, tt.expectError) {
					t.Errorf("expected error %v but got %v", tt.expectError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWishlistService_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		uniqueName    string
		quantity      int
		mockWishlist  *models.Wishlist
		wishlistError error
		updateError   error
		expectError   error
	}{
		{
			name:       "successfully update quantity",
			userID:     "user-123",
			uniqueName: "/Lotus/Item1",
			quantity:   5,
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: nil,
		},
		{
			name:        "invalid quantity zero",
			userID:      "user-123",
			uniqueName:  "/Lotus/Item1",
			quantity:    0,
			expectError: ErrInvalidQuantity,
		},
		{
			name:        "invalid quantity negative",
			userID:      "user-123",
			uniqueName:  "/Lotus/Item1",
			quantity:    -1,
			expectError: ErrInvalidQuantity,
		},
		{
			name:         "no wishlist exists",
			userID:       "user-123",
			uniqueName:   "/Lotus/Item1",
			quantity:     5,
			mockWishlist: nil,
			expectError:  ErrItemNotInWishlist,
		},
		{
			name:       "item not in wishlist",
			userID:     "user-123",
			uniqueName: "/Lotus/Item2",
			quantity:   5,
			mockWishlist: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			expectError: ErrItemNotInWishlist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishlistRepo := &mocks.MockWishlistRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
					return tt.mockWishlist, tt.wishlistError
				},
				UpdateItemQuantityFunc: func(ctx context.Context, userID, uniqueName string, quantity int) error {
					return tt.updateError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewWishlistService(mockWishlistRepo, mockItemRepo)
			err := service.UpdateQuantity(context.Background(), tt.userID, tt.uniqueName, tt.quantity)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectError)
				} else if !errors.Is(err, tt.expectError) {
					t.Errorf("expected error %v but got %v", tt.expectError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWishlistService_AddItem_WithDefaultQuantity(t *testing.T) {
	var capturedWishlist *models.Wishlist

	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			return &models.Item{UniqueName: uniqueName, Name: "Test Item"}, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return nil, nil
		},
		CreateFunc: func(ctx context.Context, wishlist *models.Wishlist) error {
			capturedWishlist = wishlist
			return nil
		},
	}

	service := NewWishlistService(mockWishlistRepo, mockItemRepo)
	err := service.AddItem(context.Background(), "user-123", models.AddItemRequest{
		UniqueName: "/Lotus/Item1",
		Quantity:   0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedWishlist == nil {
		t.Fatal("wishlist was not created")
	}

	if len(capturedWishlist.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(capturedWishlist.Items))
	}

	if capturedWishlist.Items[0].Quantity != 1 {
		t.Errorf("expected default quantity 1, got %d", capturedWishlist.Items[0].Quantity)
	}
}

func TestWishlistService_AddItem_WithTimestamp(t *testing.T) {
	var capturedItem models.WishlistItem
	beforeTest := time.Now()

	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			return &models.Item{UniqueName: uniqueName, Name: "Test Item"}, nil
		},
	}
	mockWishlistRepo := &mocks.MockWishlistRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return &models.Wishlist{UserID: userID, Items: []models.WishlistItem{}}, nil
		},
		AddItemFunc: func(ctx context.Context, userID string, item models.WishlistItem) error {
			capturedItem = item
			return nil
		},
	}

	service := NewWishlistService(mockWishlistRepo, mockItemRepo)
	err := service.AddItem(context.Background(), "user-123", models.AddItemRequest{
		UniqueName: "/Lotus/Item1",
		Quantity:   1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedItem.AddedAt.Before(beforeTest) {
		t.Error("AddedAt timestamp should be set to current time")
	}
}
