package services

import (
	"context"
	"errors"
	"testing"

	"github.com/graytonio/warframe-wishlist/internal/mocks"
	"github.com/graytonio/warframe-wishlist/internal/models"
)

func TestItemService_Search(t *testing.T) {
	tests := []struct {
		name          string
		params        models.SearchParams
		mockReturn    []models.ItemSearchResult
		mockError     error
		expectedCount int
		expectError   bool
	}{
		{
			name: "successful search with results",
			params: models.SearchParams{
				Query: "ash",
				Limit: 10,
			},
			mockReturn: []models.ItemSearchResult{
				{UniqueName: "/Lotus/Powersuits/Ninja/Ninja", Name: "Ash"},
				{UniqueName: "/Lotus/Powersuits/Ninja/NinjaPrime", Name: "Ash Prime"},
			},
			mockError:     nil,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "successful search with no results",
			params: models.SearchParams{
				Query: "nonexistent",
				Limit: 10,
			},
			mockReturn:    []models.ItemSearchResult{},
			mockError:     nil,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "search with category filter",
			params: models.SearchParams{
				Query:    "braton",
				Category: "primary",
				Limit:    10,
			},
			mockReturn: []models.ItemSearchResult{
				{UniqueName: "/Lotus/Weapons/Tenno/Rifle/Braton", Name: "Braton"},
			},
			mockError:     nil,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "search with repository error",
			params: models.SearchParams{
				Query: "error",
				Limit: 10,
			},
			mockReturn:    nil,
			mockError:     errors.New("database error"),
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockItemRepository{
				SearchFunc: func(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			service := NewItemService(mockRepo)
			results, err := service.Search(context.Background(), tt.params)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

func TestItemService_GetByUniqueName(t *testing.T) {
	tests := []struct {
		name        string
		uniqueName  string
		mockReturn  *models.Item
		mockError   error
		expectNil   bool
		expectError bool
	}{
		{
			name:       "item found",
			uniqueName: "/Lotus/Powersuits/Ninja/Ninja",
			mockReturn: &models.Item{
				UniqueName: "/Lotus/Powersuits/Ninja/Ninja",
				Name:       "Ash",
				Category:   "Warframes",
			},
			mockError:   nil,
			expectNil:   false,
			expectError: false,
		},
		{
			name:        "item not found",
			uniqueName:  "/Lotus/Nonexistent",
			mockReturn:  nil,
			mockError:   nil,
			expectNil:   true,
			expectError: false,
		},
		{
			name:        "repository error",
			uniqueName:  "/Lotus/Error",
			mockReturn:  nil,
			mockError:   errors.New("database error"),
			expectNil:   true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockItemRepository{
				FindByUniqueNameFunc: func(ctx context.Context, uniqueName string) (*models.Item, error) {
					return tt.mockReturn, tt.mockError
				},
			}

			service := NewItemService(mockRepo)
			item, err := service.GetByUniqueName(context.Background(), tt.uniqueName)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectNil && item != nil {
				t.Error("expected nil item but got value")
			}
			if !tt.expectNil && item == nil {
				t.Error("expected item but got nil")
			}
		})
	}
}
