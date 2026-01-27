package services

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
)

type ItemService struct {
	repo repository.ItemRepositoryInterface
}

func NewItemService(repo repository.ItemRepositoryInterface) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	logger.Debug(ctx, "service: ItemService.Search called", "query", params.Query, "category", params.Category)
	results, err := s.repo.Search(ctx, params)
	if err != nil {
		logger.Error(ctx, "service: ItemService.Search - repository error", "error", err)
		return nil, err
	}
	logger.Debug(ctx, "service: ItemService.Search - completed", "resultCount", len(results))
	return results, nil
}

func (s *ItemService) GetByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	logger.Debug(ctx, "service: ItemService.GetByUniqueName called", "uniqueName", uniqueName)
	item, err := s.repo.FindByUniqueName(ctx, uniqueName)
	if err != nil {
		logger.Error(ctx, "service: ItemService.GetByUniqueName - repository error", "error", err, "uniqueName", uniqueName)
		return nil, err
	}
	if item == nil {
		logger.Debug(ctx, "service: ItemService.GetByUniqueName - item not found", "uniqueName", uniqueName)
		return nil, nil
	}

	logger.Debug(ctx, "service: ItemService.GetByUniqueName - item found", "uniqueName", uniqueName, "itemName", item.Name)

	// Check which components have their own item page
	if len(item.Components) > 0 {
		componentNames := make([]string, len(item.Components))
		for i, comp := range item.Components {
			componentNames[i] = comp.UniqueName
		}

		existingItems, err := s.repo.FindByUniqueNames(ctx, componentNames)
		if err != nil {
			logger.Error(ctx, "service: ItemService.GetByUniqueName - error checking component pages", "error", err)
			// Don't fail the request, just skip populating HasOwnPage
		} else {
			for i := range item.Components {
				if _, exists := existingItems[item.Components[i].UniqueName]; exists {
					item.Components[i].HasOwnPage = true
				}
			}
		}
	}

	return item, nil
}
