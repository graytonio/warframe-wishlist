package services

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
)

type MaterialResolver struct {
	itemRepo     repository.ItemRepositoryInterface
	wishlistRepo repository.WishlistRepositoryInterface
}

func NewMaterialResolver(itemRepo repository.ItemRepositoryInterface, wishlistRepo repository.WishlistRepositoryInterface) *MaterialResolver {
	return &MaterialResolver{
		itemRepo:     itemRepo,
		wishlistRepo: wishlistRepo,
	}
}

func (r *MaterialResolver) GetMaterials(ctx context.Context, userID string) (*models.MaterialsResponse, error) {
	wishlist, err := r.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if wishlist == nil || len(wishlist.Items) == 0 {
		return &models.MaterialsResponse{
			Materials:    []models.MaterialRequirement{},
			TotalCredits: 0,
		}, nil
	}

	uniqueNames := make([]string, len(wishlist.Items))
	quantities := make(map[string]int)
	for i, item := range wishlist.Items {
		uniqueNames[i] = item.UniqueName
		quantities[item.UniqueName] = item.Quantity
	}

	items, err := r.itemRepo.FindByUniqueNames(ctx, uniqueNames)
	if err != nil {
		return nil, err
	}

	materialCounts := make(map[string]int)
	materialInfo := make(map[string]*models.Item)
	visited := make(map[string]bool)
	totalCredits := 0

	for _, wishlistItem := range wishlist.Items {
		item, exists := items[wishlistItem.UniqueName]
		if !exists {
			continue
		}

		for i := 0; i < wishlistItem.Quantity; i++ {
			for k := range visited {
				delete(visited, k)
			}
			credits := r.resolveItem(ctx, item, 1, materialCounts, materialInfo, visited)
			totalCredits += credits
		}
	}

	materials := make([]models.MaterialRequirement, 0, len(materialCounts))
	for uniqueName, count := range materialCounts {
		mat := models.MaterialRequirement{
			UniqueName: uniqueName,
			TotalCount: count,
		}

		if info, exists := materialInfo[uniqueName]; exists {
			mat.Name = info.Name
			mat.ImageName = info.ImageName
			mat.Description = info.Description
		}

		materials = append(materials, mat)
	}

	return &models.MaterialsResponse{
		Materials:    materials,
		TotalCredits: totalCredits,
	}, nil
}

func (r *MaterialResolver) resolveItem(ctx context.Context, item *models.Item, multiplier int, materialCounts map[string]int, materialInfo map[string]*models.Item, visited map[string]bool) int {
	if item == nil {
		return 0
	}

	if visited[item.UniqueName] {
		return 0
	}
	visited[item.UniqueName] = true

	totalCredits := item.BuildPrice * multiplier

	if len(item.Components) == 0 {
		materialCounts[item.UniqueName] += multiplier
		materialInfo[item.UniqueName] = item
		return totalCredits
	}

	for _, component := range item.Components {
		componentCount := component.ItemCount * multiplier

		componentItem, err := r.itemRepo.FindByUniqueName(ctx, component.UniqueName)
		if err != nil || componentItem == nil {
			materialCounts[component.UniqueName] += componentCount
			materialInfo[component.UniqueName] = &models.Item{
				UniqueName:  component.UniqueName,
				Name:        component.Name,
				ImageName:   component.ImageName,
				Description: component.Description,
			}
			continue
		}

		if len(componentItem.Components) == 0 {
			materialCounts[component.UniqueName] += componentCount
			materialInfo[component.UniqueName] = componentItem
		} else {
			credits := r.resolveItem(ctx, componentItem, componentCount, materialCounts, materialInfo, visited)
			totalCredits += credits
		}
	}

	return totalCredits
}
