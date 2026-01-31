package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/models"
)

type mockItemService struct {
	searchFunc                   func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error)
	getByUniqueNameFunc          func(ctx context.Context, uniqueName string) (*models.Item, error)
	searchReusableBlueprintsFunc func(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error)
}

func (m *mockItemService) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, params)
	}
	return nil, nil
}

func (m *mockItemService) GetByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	if m.getByUniqueNameFunc != nil {
		return m.getByUniqueNameFunc(ctx, uniqueName)
	}
	return nil, nil
}

func (m *mockItemService) SearchReusableBlueprints(ctx context.Context, query string, limit int) ([]models.ItemSearchResult, error) {
	if m.searchReusableBlueprintsFunc != nil {
		return m.searchReusableBlueprintsFunc(ctx, query, limit)
	}
	return nil, nil
}

func TestItemHandler_Search(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockReturn     []models.ItemSearchResult
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful search with results",
			queryParams: "?q=ash&limit=10",
			mockReturn: []models.ItemSearchResult{
				{UniqueName: "/Lotus/Ash", Name: "Ash"},
				{UniqueName: "/Lotus/AshPrime", Name: "Ash Prime"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "successful search with no results",
			queryParams:    "?q=nonexistent",
			mockReturn:     []models.ItemSearchResult{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "search with service error",
			queryParams:    "?q=error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "search with category filter",
			queryParams: "?q=braton&category=primary",
			mockReturn: []models.ItemSearchResult{
				{UniqueName: "/Lotus/Braton", Name: "Braton"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockItemService{
				searchFunc: func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			handler := NewItemHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/search"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			handler.Search(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				count := int(response["count"].(float64))
				if count != tt.expectedCount {
					t.Errorf("expected count %d, got %d", tt.expectedCount, count)
				}
			}
		})
	}
}

func TestItemHandler_GetByUniqueName(t *testing.T) {
	tests := []struct {
		name           string
		uniqueName     string
		mockReturn     *models.Item
		mockError      error
		expectedStatus int
	}{
		{
			name:       "item found",
			uniqueName: "Lotus-Ash",
			mockReturn: &models.Item{
				UniqueName: "/Lotus/Ash",
				Name:       "Ash",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "item not found",
			uniqueName:     "Lotus-Nonexistent",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "service error",
			uniqueName:     "Lotus-Error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockItemService{
				getByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			handler := NewItemHandler(mockService)

			r := chi.NewRouter()
			r.Get("/api/v1/items/*", handler.GetByUniqueName)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+tt.uniqueName, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestItemHandler_Search_ParsesQueryParams(t *testing.T) {
	var capturedParams models.SearchParams

	mockService := &mockItemService{
		searchFunc: func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
			capturedParams = params
			return []models.ItemSearchResult{}, nil
		},
	}

	handler := NewItemHandler(mockService)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/search?q=test&category=warframes&limit=50&offset=10", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if capturedParams.Query != "test" {
		t.Errorf("expected query 'test', got '%s'", capturedParams.Query)
	}
	if capturedParams.Category != "warframes" {
		t.Errorf("expected category 'warframes', got '%s'", capturedParams.Category)
	}
	if capturedParams.Limit != 50 {
		t.Errorf("expected limit 50, got %d", capturedParams.Limit)
	}
	if capturedParams.Offset != 10 {
		t.Errorf("expected offset 10, got %d", capturedParams.Offset)
	}
}

func TestItemHandler_GetByUniqueName_EmptyParam(t *testing.T) {
	mockService := &mockItemService{}
	handler := NewItemHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/", nil)
	rec := httptest.NewRecorder()

	handler.GetByUniqueName(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
