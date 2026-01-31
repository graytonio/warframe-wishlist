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

type mockOwnedBlueprintsService struct {
	getOwnedBlueprintsFunc func(ctx context.Context, userID string) (*models.OwnedBlueprints, error)
	addBlueprintFunc       func(ctx context.Context, userID string, req models.AddBlueprintRequest) error
	removeBlueprintFunc    func(ctx context.Context, userID, uniqueName string) error
	bulkAddBlueprintsFunc  func(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error
	clearAllBlueprintsFunc func(ctx context.Context, userID string) error
}

func (m *mockOwnedBlueprintsService) GetOwnedBlueprints(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
	if m.getOwnedBlueprintsFunc != nil {
		return m.getOwnedBlueprintsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockOwnedBlueprintsService) AddBlueprint(ctx context.Context, userID string, req models.AddBlueprintRequest) error {
	if m.addBlueprintFunc != nil {
		return m.addBlueprintFunc(ctx, userID, req)
	}
	return nil
}

func (m *mockOwnedBlueprintsService) RemoveBlueprint(ctx context.Context, userID, uniqueName string) error {
	if m.removeBlueprintFunc != nil {
		return m.removeBlueprintFunc(ctx, userID, uniqueName)
	}
	return nil
}

func (m *mockOwnedBlueprintsService) BulkAddBlueprints(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error {
	if m.bulkAddBlueprintsFunc != nil {
		return m.bulkAddBlueprintsFunc(ctx, userID, req)
	}
	return nil
}

func (m *mockOwnedBlueprintsService) ClearAllBlueprints(ctx context.Context, userID string) error {
	if m.clearAllBlueprintsFunc != nil {
		return m.clearAllBlueprintsFunc(ctx, userID)
	}
	return nil
}

func createAuthenticatedOwnedBPRequest(method, url string, body []byte, userID string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestOwnedBlueprintsHandler_GetOwnedBlueprints(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *models.OwnedBlueprints
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful get owned blueprints",
			userID: "user-123",
			mockReturn: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1", AddedAt: time.Now()},
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
			mockService := &mockOwnedBlueprintsService{
				getOwnedBlueprintsFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			handler := NewOwnedBlueprintsHandler(mockService)

			req := createAuthenticatedOwnedBPRequest(http.MethodGet, "/api/v1/profile/blueprints", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.GetOwnedBlueprints(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestOwnedBlueprintsHandler_AddBlueprint(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    models.AddBlueprintRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful add blueprint",
			userID: "user-123",
			requestBody: models.AddBlueprintRequest{
				UniqueName: "/Lotus/Blueprint1",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			requestBody:    models.AddBlueprintRequest{UniqueName: "/Lotus/Blueprint1"},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "blueprint not found",
			userID:         "user-123",
			requestBody:    models.AddBlueprintRequest{UniqueName: "/Lotus/Nonexistent"},
			mockError:      services.ErrBlueprintNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "blueprint not reusable",
			userID:         "user-123",
			requestBody:    models.AddBlueprintRequest{UniqueName: "/Lotus/Consumable"},
			mockError:      services.ErrBlueprintNotReusable,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "blueprint already owned",
			userID:         "user-123",
			requestBody:    models.AddBlueprintRequest{UniqueName: "/Lotus/Blueprint1"},
			mockError:      services.ErrBlueprintAlreadyOwned,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "missing uniqueName",
			userID:         "user-123",
			requestBody:    models.AddBlueprintRequest{UniqueName: ""},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockOwnedBlueprintsService{
				addBlueprintFunc: func(ctx context.Context, userID string, req models.AddBlueprintRequest) error {
					return tt.mockError
				},
			}

			handler := NewOwnedBlueprintsHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedOwnedBPRequest(http.MethodPost, "/api/v1/profile/blueprints", body, tt.userID)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.AddBlueprint(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestOwnedBlueprintsHandler_RemoveBlueprint(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		uniqueName     string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful remove blueprint",
			userID:         "user-123",
			uniqueName:     "Lotus/Blueprint1",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			uniqueName:     "Lotus/Blueprint1",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "blueprint not owned",
			userID:         "user-123",
			uniqueName:     "Lotus/Blueprint1",
			mockError:      services.ErrBlueprintNotOwned,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockOwnedBlueprintsService{
				removeBlueprintFunc: func(ctx context.Context, userID, uniqueName string) error {
					return tt.mockError
				},
			}

			handler := NewOwnedBlueprintsHandler(mockService)

			r := chi.NewRouter()
			r.Delete("/api/v1/profile/blueprints/*", func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.UserIDKey, tt.userID)
				handler.RemoveBlueprint(w, r.WithContext(ctx))
			})

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/profile/blueprints/"+tt.uniqueName, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestOwnedBlueprintsHandler_BulkAddBlueprints(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    models.BulkAddBlueprintsRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful bulk add",
			userID: "user-123",
			requestBody: models.BulkAddBlueprintsRequest{
				UniqueNames: []string{"/Lotus/Blueprint1", "/Lotus/Blueprint2"},
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			requestBody:    models.BulkAddBlueprintsRequest{UniqueNames: []string{"/Lotus/Blueprint1"}},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			userID:         "user-123",
			requestBody:    models.BulkAddBlueprintsRequest{UniqueNames: []string{"/Lotus/Blueprint1"}},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockOwnedBlueprintsService{
				bulkAddBlueprintsFunc: func(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error {
					return tt.mockError
				},
			}

			handler := NewOwnedBlueprintsHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := createAuthenticatedOwnedBPRequest(http.MethodPost, "/api/v1/profile/blueprints/bulk", body, tt.userID)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.BulkAddBlueprints(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestOwnedBlueprintsHandler_ClearAllBlueprints(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful clear all",
			userID:         "user-123",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user ID",
			userID:         "",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			userID:         "user-123",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockOwnedBlueprintsService{
				clearAllBlueprintsFunc: func(ctx context.Context, userID string) error {
					return tt.mockError
				},
			}

			handler := NewOwnedBlueprintsHandler(mockService)

			req := createAuthenticatedOwnedBPRequest(http.MethodDelete, "/api/v1/profile/blueprints", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.ClearAllBlueprints(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestOwnedBlueprintsHandler_AddBlueprint_InvalidJSON(t *testing.T) {
	mockService := &mockOwnedBlueprintsService{}

	handler := NewOwnedBlueprintsHandler(mockService)

	req := createAuthenticatedOwnedBPRequest(http.MethodPost, "/api/v1/profile/blueprints", []byte("invalid json"), "user-123")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.AddBlueprint(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestOwnedBlueprintsHandler_GetOwnedBlueprints_ReturnsCorrectData(t *testing.T) {
	expectedOwnedBP := &models.OwnedBlueprints{
		UserID: "user-123",
		Blueprints: []models.OwnedBlueprint{
			{UniqueName: "/Lotus/Blueprint1", AddedAt: time.Now()},
			{UniqueName: "/Lotus/Blueprint2", AddedAt: time.Now()},
		},
	}

	mockService := &mockOwnedBlueprintsService{
		getOwnedBlueprintsFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
			return expectedOwnedBP, nil
		},
	}

	handler := NewOwnedBlueprintsHandler(mockService)

	req := createAuthenticatedOwnedBPRequest(http.MethodGet, "/api/v1/profile/blueprints", nil, "user-123")
	rec := httptest.NewRecorder()

	handler.GetOwnedBlueprints(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response models.OwnedBlueprints
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.UserID != expectedOwnedBP.UserID {
		t.Errorf("expected userID '%s', got '%s'", expectedOwnedBP.UserID, response.UserID)
	}

	if len(response.Blueprints) != len(expectedOwnedBP.Blueprints) {
		t.Errorf("expected %d blueprints, got %d", len(expectedOwnedBP.Blueprints), len(response.Blueprints))
	}
}
