package services

import (
	"context"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
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
	logger.Debug(ctx, "service: MaterialResolver.GetMaterials called", "userID", userID)

	wishlist, err := r.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: MaterialResolver.GetMaterials - error fetching wishlist", "error", err)
		return nil, err
	}

	if wishlist == nil || len(wishlist.Items) == 0 {
		logger.Debug(ctx, "service: MaterialResolver.GetMaterials - empty wishlist, returning empty materials")
		return &models.MaterialsResponse{
			Materials:    []models.MaterialRequirement{},
			TotalCredits: 0,
		}, nil
	}

	logger.Debug(ctx, "service: MaterialResolver.GetMaterials - processing wishlist items", "itemCount", len(wishlist.Items))

	uniqueNames := make([]string, len(wishlist.Items))
	quantities := make(map[string]int)
	for i, item := range wishlist.Items {
		uniqueNames[i] = item.UniqueName
		quantities[item.UniqueName] = item.Quantity
	}

	logger.Debug(ctx, "service: MaterialResolver.GetMaterials - fetching item details")
	items, err := r.itemRepo.FindByUniqueNames(ctx, uniqueNames)
	if err != nil {
		logger.Error(ctx, "service: MaterialResolver.GetMaterials - error fetching items", "error", err)
		return nil, err
	}
	logger.Debug(ctx, "service: MaterialResolver.GetMaterials - fetched item details", "foundCount", len(items))

	materialCounts := make(map[string]int)
	materialInfo := make(map[string]*models.Item)
	visited := make(map[string]bool)
	totalCredits := 0

	for _, wishlistItem := range wishlist.Items {
		item, exists := items[wishlistItem.UniqueName]
		if !exists {
			logger.Debug(ctx, "service: MaterialResolver.GetMaterials - item not found in database, skipping", "uniqueName", wishlistItem.UniqueName)
			continue
		}

		logger.Debug(ctx, "service: MaterialResolver.GetMaterials - resolving materials for item", "uniqueName", wishlistItem.UniqueName, "quantity", wishlistItem.Quantity)
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

	logger.Info(ctx, "service: MaterialResolver.GetMaterials - completed", "materialCount", len(materials), "totalCredits", totalCredits)
	return &models.MaterialsResponse{
		Materials:    materials,
		TotalCredits: totalCredits,
	}, nil
}

func (r *MaterialResolver) resolveItem(ctx context.Context, item *models.Item, multiplier int, materialCounts map[string]int, materialInfo map[string]*models.Item, visited map[string]bool) int {
	if item == nil {
		logger.Debug(ctx, "service: MaterialResolver.resolveItem - nil item, returning 0")
		return 0
	}

	if visited[item.UniqueName] {
		logger.Debug(ctx, "service: MaterialResolver.resolveItem - already visited, skipping", "uniqueName", item.UniqueName)
		return 0
	}
	visited[item.UniqueName] = true

	totalCredits := item.BuildPrice * multiplier
	logger.Debug(ctx, "service: MaterialResolver.resolveItem - processing", "uniqueName", item.UniqueName, "multiplier", multiplier, "buildPrice", item.BuildPrice)

	if len(item.Components) == 0 {
		logger.Debug(ctx, "service: MaterialResolver.resolveItem - base material (no components)", "uniqueName", item.UniqueName, "count", multiplier)
		materialCounts[item.UniqueName] += multiplier
		materialInfo[item.UniqueName] = item
		return totalCredits
	}

	logger.Debug(ctx, "service: MaterialResolver.resolveItem - processing components", "uniqueName", item.UniqueName, "componentCount", len(item.Components))
	for _, component := range item.Components {
		componentCount := component.ItemCount * multiplier

		componentItem, err := r.itemRepo.FindByUniqueName(ctx, component.UniqueName)
		if err != nil || componentItem == nil {
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - component not found or error, treating as raw material", "uniqueName", component.UniqueName, "count", componentCount)
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
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - component is base material", "uniqueName", component.UniqueName, "count", componentCount)
			materialCounts[component.UniqueName] += componentCount
			materialInfo[component.UniqueName] = componentItem
		} else {
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - recursing into component", "uniqueName", component.UniqueName)
			credits := r.resolveItem(ctx, componentItem, componentCount, materialCounts, materialInfo, visited)
			totalCredits += credits
		}
	}

	return totalCredits
}
