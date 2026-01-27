package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/middleware"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/services"
)

type mockWishlistService struct {
	getWishlistFunc    func(ctx context.Context, userID string) (*models.Wishlist, error)
	addItemFunc        func(ctx context.Context, userID string, req models.AddItemRequest) error
	removeItemFunc     func(ctx context.Context, userID, uniqueName string) error
	updateQuantityFunc func(ctx context.Context, userID, uniqueName string, quantity int) error
}

func (m *mockWishlistService) GetWishlist(ctx context.Context, userID string) (*models.Wishlist, error) {
	if m.getWishlistFunc != nil {
		return m.getWishlistFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockWishlistService) AddItem(ctx context.Context, userID string, req models.AddItemRequest) error {
	if m.addItemFunc != nil {
		return m.addItemFunc(ctx, userID, req)
	}
	return nil
}

func (m *mockWishlistService) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	if m.removeItemFunc != nil {
		return m.removeItemFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *mockWishlistService) UpdateQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	if m.updateQuantityFunc != nil {
		return m.updateQuantityFunc(ctx, userID, uniqueName, quantity)
	}
	return nil
}

type mockMaterialResolver struct {
	getMaterialsFunc func(ctx context.Context, userID string) (*models.MaterialsResponse, error)
}

func (m *mockMaterialResolver) GetMaterials(ctx context.Context, userID string) (*models.MaterialsResponse, error) {
	if m.getMaterialsFunc != nil {
		return m.getMaterialsFunc(ctx, userID)
	}
	return nil, nil
}

func createAuthenticatedRequest(method, url string, body []byte, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestWishlistHandler_GetWishlist(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *models.Wishlist
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful get wishlist",
			userID: "user-123",
			mockReturn: &models.Wishlist{
				UserID: "user-123",
				Items: []models.WishlistItem{
					{UniqueName: "/Lotus/Item1", Quantity: 1},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			userID:         "user-123",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockWishlistService{
				getWishlistFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
					return tt.mockReturn, tt.mockError
				},
			}
			mockResolver := &mockMaterialResolver{}

			handler := NewWishlistHandler(mockService, mockResolver)

			req := createAuthenticatedRequest(http.MethodGet, "/api/v1/wishlist", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.GetWishlist(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestWishlistHandler_AddItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    models.AddItemRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful add item",
			userID: "user-123",
			requestBody: models.AddItemRequest{
				UniqueName: "/Lotus/Item1",
				Quantity:   1,
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			requestBody:    models.AddItemRequest{UniqueName: "/Lotus/Item1"},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "item not found",
			userID:         "user-123",
			requestBody:    models.AddItemRequest{UniqueName: "/Lotus/Nonexistent"},
			mockError:      services.ErrItemNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "item already in wishlist",
			userID:         "user-123",
			requestBody:    models.AddItemRequest{UniqueName: "/Lotus/Item1"},
			mockError:      services.ErrItemAlreadyInWishlist,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "missing uniqueName",
			userID:         "user-123",
			requestBody:    models.AddItemRequest{UniqueName: ""},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockWishlistService{
				addItemFunc: func(ctx context.Context, userID string, req models.AddItemRequest) error {
					return tt.mockError
				},
			}
			mockResolver := &mockMaterialResolver{}

			handler := NewWishlistHandler(mockService, mockResolver)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedRequest(http.MethodPost, "/api/v1/wishlist", body, tt.userID)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.AddItem(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestWishlistHandler_RemoveItem(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		uniqueName     string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful remove item",
			userID:         "user-123",
			uniqueName:     "Lotus-Item1",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			uniqueName:     "Lotus-Item1",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "item not in wishlist",
			userID:         "user-123",
			uniqueName:     "Lotus-Item1",
			mockError:      services.ErrItemNotInWishlist,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockWishlistService{
				removeItemFunc: func(ctx context.Context, userID, uniqueName string) error {
					return tt.mockError
				},
			}
			mockResolver := &mockMaterialResolver{}

			handler := NewWishlistHandler(mockService, mockResolver)

			r := chi.NewRouter()
			r.Delete("/api/v1/wishlist/{uniqueName}", func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.UserIDKey, tt.userID)
				handler.RemoveItem(w, r.WithContext(ctx))
			})

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/"+tt.uniqueName, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestWishlistHandler_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		uniqueName     string
		requestBody    models.UpdateQuantityRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful update quantity",
			userID:         "user-123",
			uniqueName:     "Lotus-Item1",
			requestBody:    models.UpdateQuantityRequest{Quantity: 5},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			uniqueName:     "Lotus-Item1",
			requestBody:    models.UpdateQuantityRequest{Quantity: 5},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "item not in wishlist",
			userID:         "user-123",
			uniqueName:     "Lotus-Item1",
			requestBody:    models.UpdateQuantityRequest{Quantity: 5},
			mockError:      services.ErrItemNotInWishlist,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid quantity",
			userID:         "user-123",
			uniqueName:     "Lotus-Item1",
			requestBody:    models.UpdateQuantityRequest{Quantity: 0},
			mockError:      services.ErrInvalidQuantity,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockWishlistService{
				updateQuantityFunc: func(ctx context.Context, userID, uniqueName string, quantity int) error {
					return tt.mockError
				},
			}
			mockResolver := &mockMaterialResolver{}

			handler := NewWishlistHandler(mockService, mockResolver)

			r := chi.NewRouter()
			r.Patch("/api/v1/wishlist/{uniqueName}", func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.UserIDKey, tt.userID)
				handler.UpdateQuantity(w, r.WithContext(ctx))
			})

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/wishlist/"+tt.uniqueName, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestWishlistHandler_GetMaterials(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *models.MaterialsResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful get materials",
			userID: "user-123",
			mockReturn: &models.MaterialsResponse{
				Materials: []models.MaterialRequirement{
					{UniqueName: "/Lotus/Resource1", Name: "Resource 1", TotalCount: 100},
				},
				TotalCredits: 25000,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			userID:         "user-123",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockWishlistService{}
			mockResolver := &mockMaterialResolver{
				getMaterialsFunc: func(ctx context.Context, userID string) (*models.MaterialsResponse, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			handler := NewWishlistHandler(mockService, mockResolver)

			req := createAuthenticatedRequest(http.MethodGet, "/api/v1/wishlist/materials", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.GetMaterials(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestWishlistHandler_AddItem_InvalidJSON(t *testing.T) {
	mockService := &mockWishlistService{}
	mockResolver := &mockMaterialResolver{}

	handler := NewWishlistHandler(mockService, mockResolver)

	req := createAuthenticatedRequest(http.MethodPost, "/api/v1/wishlist", []byte("invalid json"), "user-123")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.AddItem(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestWishlistHandler_GetWishlist_ReturnsCorrectData(t *testing.T) {
	expectedWishlist := &models.Wishlist{
		UserID: "user-123",
		Items: []models.WishlistItem{
			{UniqueName: "/Lotus/Item1", Quantity: 2, AddedAt: time.Now()},
			{UniqueName: "/Lotus/Item2", Quantity: 1, AddedAt: time.Now()},
		},
	}

	mockService := &mockWishlistService{
		getWishlistFunc: func(ctx context.Context, userID string) (*models.Wishlist, error) {
			return expectedWishlist, nil
		},
	}
	mockResolver := &mockMaterialResolver{}

	handler := NewWishlistHandler(mockService, mockResolver)

	req := createAuthenticatedRequest(http.MethodGet, "/api/v1/wishlist", nil, "user-123")
	rec := httptest.NewRecorder()

	handler.GetWishlist(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response models.Wishlist
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.UserID != expectedWishlist.UserID {
		t.Errorf("expected userID '%s', got '%s'", expectedWishlist.UserID, response.UserID)
	}

	if len(response.Items) != len(expectedWishlist.Items) {
		t.Errorf("expected %d items, got %d", len(expectedWishlist.Items), len(response.Items))
	}
}
