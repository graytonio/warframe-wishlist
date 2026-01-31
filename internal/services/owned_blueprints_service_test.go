package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/mocks"
	"github.com/graytonio/warframe-wishlist/internal/models"
)

func TestOwnedBlueprintsService_GetOwnedBlueprints(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *models.OwnedBlueprints
		mockError      error
		expectError    bool
		expectNewEmpty bool
	}{
		{
			name:   "existing owned blueprints found",
			userID: "user-123",
			mockReturn: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1", AddedAt: time.Now()},
				},
			},
			mockError:      nil,
			expectError:    false,
			expectNewEmpty: false,
		},
		{
			name:           "no owned blueprints returns empty",
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
			mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockReturn, tt.mockError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
			result, err := service.GetOwnedBlueprints(context.Background(), tt.userID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result == nil {
				t.Error("expected result but got nil")
			}
			if tt.expectNewEmpty && result != nil && len(result.Blueprints) != 0 {
				t.Errorf("expected empty blueprints, got %d", len(result.Blueprints))
			}
		})
	}
}

func TestOwnedBlueprintsService_AddBlueprint(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		request        models.AddBlueprintRequest
		mockItem       *models.Item
		mockOwnedBP    *models.OwnedBlueprints
		itemError      error
		ownedBPError   error
		createError    error
		addError       error
		expectError    error
	}{
		{
			name:   "add blueprint to new user",
			userID: "user-123",
			request: models.AddBlueprintRequest{
				UniqueName: "/Lotus/Blueprint1",
			},
			mockItem:     &models.Item{UniqueName: "/Lotus/Blueprint1", Name: "Blueprint 1", ConsumeOnBuild: false},
			mockOwnedBP:  nil,
			expectError:  nil,
		},
		{
			name:   "add blueprint to existing user",
			userID: "user-123",
			request: models.AddBlueprintRequest{
				UniqueName: "/Lotus/Blueprint2",
			},
			mockItem: &models.Item{UniqueName: "/Lotus/Blueprint2", Name: "Blueprint 2", ConsumeOnBuild: false},
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
				},
			},
			expectError: nil,
		},
		{
			name:   "blueprint not found",
			userID: "user-123",
			request: models.AddBlueprintRequest{
				UniqueName: "/Lotus/Nonexistent",
			},
			mockItem:    nil,
			expectError: ErrBlueprintNotFound,
		},
		{
			name:   "blueprint not reusable",
			userID: "user-123",
			request: models.AddBlueprintRequest{
				UniqueName: "/Lotus/ConsumableBlueprint",
			},
			mockItem:    &models.Item{UniqueName: "/Lotus/ConsumableBlueprint", Name: "Consumable", ConsumeOnBuild: true},
			expectError: ErrBlueprintNotReusable,
		},
		{
			name:   "blueprint already owned",
			userID: "user-123",
			request: models.AddBlueprintRequest{
				UniqueName: "/Lotus/Blueprint1",
			},
			mockItem: &models.Item{UniqueName: "/Lotus/Blueprint1", Name: "Blueprint 1", ConsumeOnBuild: false},
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
				},
			},
			expectError: ErrBlueprintAlreadyOwned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockItemRepo := &mocks.MockItemRepository{
				FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
					return tt.mockItem, tt.itemError
				},
			}
			mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockOwnedBP, tt.ownedBPError
				},
				CreateFunc: func(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error {
					return tt.createError
				},
				AddBlueprintFunc: func(ctx context.Context, userID string, blueprint models.OwnedBlueprint) error {
					return tt.addError
				},
			}

			service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
			err := service.AddBlueprint(context.Background(), tt.userID, tt.request)

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

func TestOwnedBlueprintsService_RemoveBlueprint(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		uniqueName   string
		mockOwnedBP  *models.OwnedBlueprints
		ownedBPError error
		removeError  error
		expectError  error
	}{
		{
			name:       "successfully remove blueprint",
			userID:     "user-123",
			uniqueName: "/Lotus/Blueprint1",
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
				},
			},
			expectError: nil,
		},
		{
			name:        "no owned blueprints exists",
			userID:      "user-123",
			uniqueName:  "/Lotus/Blueprint1",
			mockOwnedBP: nil,
			expectError: ErrBlueprintNotOwned,
		},
		{
			name:       "blueprint not owned",
			userID:     "user-123",
			uniqueName: "/Lotus/Blueprint2",
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
				},
			},
			expectError: ErrBlueprintNotOwned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockOwnedBP, tt.ownedBPError
				},
				RemoveBlueprintFunc: func(ctx context.Context, userID, uniqueName string) error {
					return tt.removeError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
			err := service.RemoveBlueprint(context.Background(), tt.userID, tt.uniqueName)

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

func TestOwnedBlueprintsService_BulkAddBlueprints(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		request     models.BulkAddBlueprintsRequest
		mockItems   map[string]*models.Item
		mockOwnedBP *models.OwnedBlueprints
		expectError bool
	}{
		{
			name:   "bulk add new blueprints",
			userID: "user-123",
			request: models.BulkAddBlueprintsRequest{
				UniqueNames: []string{"/Lotus/Blueprint1", "/Lotus/Blueprint2"},
			},
			mockItems: map[string]*models.Item{
				"/Lotus/Blueprint1": {UniqueName: "/Lotus/Blueprint1", ConsumeOnBuild: false},
				"/Lotus/Blueprint2": {UniqueName: "/Lotus/Blueprint2", ConsumeOnBuild: false},
			},
			mockOwnedBP: nil,
			expectError: false,
		},
		{
			name:   "skip already owned blueprints",
			userID: "user-123",
			request: models.BulkAddBlueprintsRequest{
				UniqueNames: []string{"/Lotus/Blueprint1", "/Lotus/Blueprint2"},
			},
			mockItems: map[string]*models.Item{
				"/Lotus/Blueprint1": {UniqueName: "/Lotus/Blueprint1", ConsumeOnBuild: false},
				"/Lotus/Blueprint2": {UniqueName: "/Lotus/Blueprint2", ConsumeOnBuild: false},
			},
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
				},
			},
			expectError: false,
		},
		{
			name:   "skip consumable blueprints",
			userID: "user-123",
			request: models.BulkAddBlueprintsRequest{
				UniqueNames: []string{"/Lotus/Consumable", "/Lotus/Reusable"},
			},
			mockItems: map[string]*models.Item{
				"/Lotus/Consumable": {UniqueName: "/Lotus/Consumable", ConsumeOnBuild: true},
				"/Lotus/Reusable":   {UniqueName: "/Lotus/Reusable", ConsumeOnBuild: false},
			},
			mockOwnedBP: nil,
			expectError: false,
		},
		{
			name:   "empty request",
			userID: "user-123",
			request: models.BulkAddBlueprintsRequest{
				UniqueNames: []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockItemRepo := &mocks.MockItemRepository{
				FindByUniqueNamesFunc: func(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
					return tt.mockItems, nil
				},
			}
			mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockOwnedBP, nil
				},
				CreateFunc: func(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error {
					return nil
				},
				BulkAddBlueprintsFunc: func(ctx context.Context, userID string, blueprints []models.OwnedBlueprint) error {
					return nil
				},
			}

			service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
			err := service.BulkAddBlueprints(context.Background(), tt.userID, tt.request)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOwnedBlueprintsService_ClearAllBlueprints(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		mockOwnedBP  *models.OwnedBlueprints
		ownedBPError error
		clearError   error
		expectError  bool
	}{
		{
			name:   "successfully clear all blueprints",
			userID: "user-123",
			mockOwnedBP: &models.OwnedBlueprints{
				UserID: "user-123",
				Blueprints: []models.OwnedBlueprint{
					{UniqueName: "/Lotus/Blueprint1"},
					{UniqueName: "/Lotus/Blueprint2"},
				},
			},
			expectError: false,
		},
		{
			name:        "no owned blueprints to clear",
			userID:      "user-123",
			mockOwnedBP: nil,
			expectError: false,
		},
		{
			name:         "repository error",
			userID:       "user-123",
			ownedBPError: errors.New("database error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
				GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
					return tt.mockOwnedBP, tt.ownedBPError
				},
				ClearAllFunc: func(ctx context.Context, userID string) error {
					return tt.clearError
				},
			}
			mockItemRepo := &mocks.MockItemRepository{}

			service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
			err := service.ClearAllBlueprints(context.Background(), tt.userID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOwnedBlueprintsService_AddBlueprint_WithTimestamp(t *testing.T) {
	var capturedOwnedBP *models.OwnedBlueprints
	beforeTest := time.Now()

	mockItemRepo := &mocks.MockItemRepository{
		FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
			return &models.Item{UniqueName: uniqueName, Name: "Test Blueprint", ConsumeOnBuild: false}, nil
		},
	}
	mockOwnedBPRepo := &mocks.MockOwnedBlueprintsRepository{
		GetByUserIDFunc: func(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
			return nil, nil
		},
		CreateFunc: func(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error {
			capturedOwnedBP = ownedBlueprints
			return nil
		},
	}

	service := NewOwnedBlueprintsService(mockOwnedBPRepo, mockItemRepo)
	err := service.AddBlueprint(context.Background(), "user-123", models.AddBlueprintRequest{
		UniqueName: "/Lotus/Blueprint1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedOwnedBP == nil {
		t.Fatal("owned blueprints was not created")
	}

	if len(capturedOwnedBP.Blueprints) != 1 {
		t.Fatalf("expected 1 blueprint, got %d", len(capturedOwnedBP.Blueprints))
	}

	if capturedOwnedBP.Blueprints[0].AddedAt.Before(beforeTest) {
		t.Error("AddedAt timestamp should be set to current time")
	}
}
