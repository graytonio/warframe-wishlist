package services

import (
	"context"
	"strings"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
)

type MaterialResolver struct {
	itemRepo     repository.ItemRepositoryInterface
	wishlistRepo repository.WishlistRepositoryInterface
	ownedBPRepo  repository.OwnedBlueprintsRepositoryInterface
}

func NewMaterialResolver(itemRepo repository.ItemRepositoryInterface, wishlistRepo repository.WishlistRepositoryInterface, ownedBPRepo repository.OwnedBlueprintsRepositoryInterface) *MaterialResolver {
	return &MaterialResolver{
		itemRepo:     itemRepo,
		wishlistRepo: wishlistRepo,
		ownedBPRepo:  ownedBPRepo,
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

	// Fetch owned blueprints to exclude from materials
	ownedBlueprintsSet := make(map[string]bool)
	if r.ownedBPRepo != nil {
		ownedBP, err := r.ownedBPRepo.GetByUserID(ctx, userID)
		if err != nil {
			logger.Error(ctx, "service: MaterialResolver.GetMaterials - error fetching owned blueprints", "error", err)
			return nil, err
		}
		if ownedBP != nil {
			for _, bp := range ownedBP.Blueprints {
				ownedBlueprintsSet[bp.UniqueName] = true
			}
			logger.Debug(ctx, "service: MaterialResolver.GetMaterials - fetched owned blueprints", "count", len(ownedBP.Blueprints))
		}
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
	nonConsumableCounted := make(map[string]bool) // Track non-consumable items globally
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
			credits := r.resolveItemInternal(ctx, item, "", 1, materialCounts, materialInfo, visited, nonConsumableCounted, ownedBlueprintsSet)
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
	nonConsumableCounted := make(map[string]bool)
	ownedBlueprintsSet := make(map[string]bool)
	return r.resolveItemInternal(ctx, item, "", multiplier, materialCounts, materialInfo, visited, nonConsumableCounted, ownedBlueprintsSet)
}

// ceilDiv performs ceiling division: ceil(a / b)
func ceilDiv(a, b int) int {
	if b <= 0 {
		return a
	}
	return (a + b - 1) / b
}

// isLikelyBlueprint determines if an item is a blueprint type that should be treated as reusable.
// Returns true if the item appears to be a blueprint (has "Blueprint" in the name or unique name).
// Regular consumable materials should return false.
func isLikelyBlueprint(item *models.Item) bool {
	if item == nil {
		return false
	}
	// Check if "Blueprint" appears in the name or unique name
	// This helps distinguish blueprints from regular consumable materials
	return containsIgnoreCase(item.Name, "Blueprint") || containsIgnoreCase(item.UniqueName, "Blueprint")
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || strings.Contains(s, substr)))
}

func (r *MaterialResolver) resolveItemInternal(ctx context.Context, item *models.Item, parentName string, multiplier int, materialCounts map[string]int, materialInfo map[string]*models.Item, visited map[string]bool, nonConsumableCounted map[string]bool, ownedBlueprintsSet map[string]bool) int {
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
		// Determine if this is actually a reusable blueprint
		// ConsumeOnBuild defaults to false in Go, so we need additional checks
		// A reusable blueprint must have ConsumeOnBuild=false AND be a blueprint-type item
		isReusableBlueprint := !item.ConsumeOnBuild && isLikelyBlueprint(item)

		// Check if this is a reusable blueprint that user already owns
		if isReusableBlueprint && ownedBlueprintsSet[item.UniqueName] {
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - user already owns this reusable blueprint, skipping", "uniqueName", item.UniqueName)
			return totalCredits
		}

		// Check if this is a reusable blueprint already counted
		if isReusableBlueprint && nonConsumableCounted[item.UniqueName] {
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - non-consumable already counted, skipping", "uniqueName", item.UniqueName)
			return totalCredits
		}

		countToAdd := multiplier
		if isReusableBlueprint {
			// Non-consumable items only need 1 regardless of quantity
			countToAdd = 1
			nonConsumableCounted[item.UniqueName] = true
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - non-consumable base material", "uniqueName", item.UniqueName)
		} else {
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - base material (no components)", "uniqueName", item.UniqueName, "count", multiplier)
		}

		materialCounts[item.UniqueName] += countToAdd
		// For items named "Blueprint", add parent context
		itemToStore := item
		if item.Name == "Blueprint" && parentName != "" {
			itemToStore = &models.Item{
				UniqueName:  item.UniqueName,
				Name:        "Blueprint (" + parentName + ")",
				ImageName:   item.ImageName,
				Description: item.Description,
			}
		}
		materialInfo[item.UniqueName] = itemToStore
		return totalCredits
	}

	logger.Debug(ctx, "service: MaterialResolver.resolveItem - processing components", "uniqueName", item.UniqueName, "componentCount", len(item.Components))
	for _, component := range item.Components {
		componentCount := component.ItemCount * multiplier

		// Check if component has nested components in the embedded data
		if len(component.Components) > 0 {
			// Try to fetch from database to get buildQuantity
			componentItem, _ := r.itemRepo.FindByUniqueName(ctx, component.UniqueName)
			buildQuantity := 1
			if componentItem != nil && componentItem.BuildQuantity > 0 {
				buildQuantity = componentItem.BuildQuantity
			}
			craftsNeeded := ceilDiv(componentCount, buildQuantity)
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - component has nested components, recursing", "uniqueName", component.UniqueName, "needed", componentCount, "buildQuantity", buildQuantity, "crafts", craftsNeeded)
			// Create a temporary Item from the component to recurse
			componentAsItem := &models.Item{
				UniqueName:  component.UniqueName,
				Name:        component.Name,
				ImageName:   component.ImageName,
				Description: component.Description,
				Components:  component.Components,
			}
			credits := r.resolveItemInternal(ctx, componentAsItem, item.Name, craftsNeeded, materialCounts, materialInfo, visited, nonConsumableCounted, ownedBlueprintsSet)
			totalCredits += credits
			continue
		}

		// Try to fetch from database to check for additional components
		componentItem, err := r.itemRepo.FindByUniqueName(ctx, component.UniqueName)
		if err != nil || componentItem == nil {
			// Component not found in database and has no nested components - it's a base material
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - component is base material (not in db)", "uniqueName", component.UniqueName, "count", componentCount)
			materialCounts[component.UniqueName] += componentCount
			// For components named "Blueprint", add parent context
			componentName := component.Name
			if component.Name == "Blueprint" && item.Name != "" {
				componentName = "Blueprint (" + item.Name + ")"
			}
			materialInfo[component.UniqueName] = &models.Item{
				UniqueName:  component.UniqueName,
				Name:        componentName,
				ImageName:   component.ImageName,
				Description: component.Description,
			}
			continue
		}

		if len(componentItem.Components) == 0 {
			// Determine if this is actually a reusable blueprint
			// A reusable blueprint must have ConsumeOnBuild=false AND be a blueprint-type item
			isReusableBlueprint := !componentItem.ConsumeOnBuild && isLikelyBlueprint(componentItem)

			// Check if this is a reusable blueprint that user already owns
			if isReusableBlueprint && ownedBlueprintsSet[component.UniqueName] {
				logger.Debug(ctx, "service: MaterialResolver.resolveItem - user already owns this reusable blueprint, skipping", "uniqueName", component.UniqueName)
				continue
			}

			// Check if this is a reusable blueprint already counted
			if isReusableBlueprint && nonConsumableCounted[component.UniqueName] {
				logger.Debug(ctx, "service: MaterialResolver.resolveItem - non-consumable already counted, skipping", "uniqueName", component.UniqueName)
				continue
			}

			countToAdd := componentCount
			if isReusableBlueprint {
				// Non-consumable items only need 1 regardless of quantity
				countToAdd = 1
				nonConsumableCounted[component.UniqueName] = true
				logger.Debug(ctx, "service: MaterialResolver.resolveItem - non-consumable component", "uniqueName", component.UniqueName)
			} else {
				logger.Debug(ctx, "service: MaterialResolver.resolveItem - component is base material", "uniqueName", component.UniqueName, "count", componentCount)
			}

			materialCounts[component.UniqueName] += countToAdd
			// For components named "Blueprint", add parent context
			if componentItem.Name == "Blueprint" && item.Name != "" {
				componentItem = &models.Item{
					UniqueName:  componentItem.UniqueName,
					Name:        "Blueprint (" + item.Name + ")",
					ImageName:   componentItem.ImageName,
					Description: componentItem.Description,
				}
			}
			materialInfo[component.UniqueName] = componentItem
		} else {
			// Calculate crafts needed based on buildQuantity
			buildQuantity := 1
			if componentItem.BuildQuantity > 0 {
				buildQuantity = componentItem.BuildQuantity
			}
			craftsNeeded := ceilDiv(componentCount, buildQuantity)
			logger.Debug(ctx, "service: MaterialResolver.resolveItem - recursing into component", "uniqueName", component.UniqueName, "needed", componentCount, "buildQuantity", buildQuantity, "crafts", craftsNeeded)
			credits := r.resolveItemInternal(ctx, componentItem, item.Name, craftsNeeded, materialCounts, materialInfo, visited, nonConsumableCounted, ownedBlueprintsSet)
			totalCredits += credits
		}
	}

	return totalCredits
}
