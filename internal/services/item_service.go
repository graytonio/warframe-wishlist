package services

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
)

type ItemService struct {
	repo repository.ItemRepositoryInterface
}

func NewItemService(repo repository.ItemRepositoryInterface) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	return s.repo.Search(ctx, params)
}

func (s *ItemService) GetByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	return s.repo.FindByUniqueName(ctx, uniqueName)
}
