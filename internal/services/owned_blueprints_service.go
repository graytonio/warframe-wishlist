package services

import (
	"context"
	"errors"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
)

var (
	ErrBlueprintNotFound       = errors.New("blueprint not found")
	ErrBlueprintNotReusable    = errors.New("blueprint is not reusable (consumeOnBuild is true)")
	ErrBlueprintAlreadyOwned   = errors.New("blueprint already owned")
	ErrBlueprintNotOwned       = errors.New("blueprint not owned")
)

type OwnedBlueprintsService struct {
	ownedBPRepo repository.OwnedBlueprintsRepositoryInterface
	itemRepo    repository.ItemRepositoryInterface
}

func NewOwnedBlueprintsService(ownedBPRepo repository.OwnedBlueprintsRepositoryInterface, itemRepo repository.ItemRepositoryInterface) *OwnedBlueprintsService {
	return &OwnedBlueprintsService{
		ownedBPRepo: ownedBPRepo,
		itemRepo:    itemRepo,
	}
}

func (s *OwnedBlueprintsService) GetOwnedBlueprints(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
	logger.Debug(ctx, "service: OwnedBlueprintsService.GetOwnedBlueprints called", "userID", userID)

	ownedBP, err := s.ownedBPRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.GetOwnedBlueprints - repository error", "error", err)
		return nil, err
	}

	if ownedBP == nil {
		logger.Debug(ctx, "service: OwnedBlueprintsService.GetOwnedBlueprints - creating empty owned blueprints for new user")
		ownedBP = &models.OwnedBlueprints{
			UserID:     userID,
			Blueprints: []models.OwnedBlueprint{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}

	logger.Debug(ctx, "service: OwnedBlueprintsService.GetOwnedBlueprints - completed", "blueprintCount", len(ownedBP.Blueprints))
	return ownedBP, nil
}

func (s *OwnedBlueprintsService) AddBlueprint(ctx context.Context, userID string, req models.AddBlueprintRequest) error {
	logger.Debug(ctx, "service: OwnedBlueprintsService.AddBlueprint called", "userID", userID, "uniqueName", req.UniqueName)

	// Validate item exists and is reusable
	item, err := s.itemRepo.FindByUniqueName(ctx, req.UniqueName)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.AddBlueprint - error finding item", "error", err)
		return err
	}
	if item == nil {
		logger.Warn(ctx, "service: OwnedBlueprintsService.AddBlueprint - item not found", "uniqueName", req.UniqueName)
		return ErrBlueprintNotFound
	}
	if item.ConsumeOnBuild {
		logger.Warn(ctx, "service: OwnedBlueprintsService.AddBlueprint - blueprint is not reusable", "uniqueName", req.UniqueName)
		return ErrBlueprintNotReusable
	}

	// Get or create owned blueprints
	ownedBP, err := s.ownedBPRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.AddBlueprint - error fetching owned blueprints", "error", err)
		return err
	}

	if ownedBP == nil {
		// Create new owned blueprints document
		logger.Debug(ctx, "service: OwnedBlueprintsService.AddBlueprint - creating new owned blueprints for user")
		ownedBP = &models.OwnedBlueprints{
			UserID: userID,
			Blueprints: []models.OwnedBlueprint{
				{
					UniqueName: req.UniqueName,
					AddedAt:    time.Now(),
				},
			},
		}
		err = s.ownedBPRepo.Create(ctx, ownedBP)
		if err != nil {
			logger.Error(ctx, "service: OwnedBlueprintsService.AddBlueprint - error creating owned blueprints", "error", err)
			return err
		}
		logger.Info(ctx, "service: OwnedBlueprintsService.AddBlueprint - created new owned blueprints with blueprint", "uniqueName", req.UniqueName)
		return nil
	}

	// Check for duplicates
	for _, bp := range ownedBP.Blueprints {
		if bp.UniqueName == req.UniqueName {
			logger.Warn(ctx, "service: OwnedBlueprintsService.AddBlueprint - blueprint already owned", "uniqueName", req.UniqueName)
			return ErrBlueprintAlreadyOwned
		}
	}

	// Add blueprint
	newBlueprint := models.OwnedBlueprint{
		UniqueName: req.UniqueName,
		AddedAt:    time.Now(),
	}

	err = s.ownedBPRepo.AddBlueprint(ctx, userID, newBlueprint)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.AddBlueprint - error adding blueprint", "error", err)
		return err
	}

	logger.Info(ctx, "service: OwnedBlueprintsService.AddBlueprint - blueprint added successfully", "uniqueName", req.UniqueName)
	return nil
}

func (s *OwnedBlueprintsService) RemoveBlueprint(ctx context.Context, userID, uniqueName string) error {
	logger.Debug(ctx, "service: OwnedBlueprintsService.RemoveBlueprint called", "userID", userID, "uniqueName", uniqueName)

	ownedBP, err := s.ownedBPRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.RemoveBlueprint - error fetching owned blueprints", "error", err)
		return err
	}

	if ownedBP == nil {
		logger.Warn(ctx, "service: OwnedBlueprintsService.RemoveBlueprint - no owned blueprints found for user")
		return ErrBlueprintNotOwned
	}

	// Check if blueprint is owned
	found := false
	for _, bp := range ownedBP.Blueprints {
		if bp.UniqueName == uniqueName {
			found = true
			break
		}
	}

	if !found {
		logger.Warn(ctx, "service: OwnedBlueprintsService.RemoveBlueprint - blueprint not owned", "uniqueName", uniqueName)
		return ErrBlueprintNotOwned
	}

	err = s.ownedBPRepo.RemoveBlueprint(ctx, userID, uniqueName)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.RemoveBlueprint - error removing blueprint", "error", err)
		return err
	}

	logger.Info(ctx, "service: OwnedBlueprintsService.RemoveBlueprint - blueprint removed successfully", "uniqueName", uniqueName)
	return nil
}

func (s *OwnedBlueprintsService) BulkAddBlueprints(ctx context.Context, userID string, req models.BulkAddBlueprintsRequest) error {
	logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints called", "userID", userID, "count", len(req.UniqueNames))

	if len(req.UniqueNames) == 0 {
		logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - empty request, nothing to do")
		return nil
	}

	// Validate all items exist and are reusable
	items, err := s.itemRepo.FindByUniqueNames(ctx, req.UniqueNames)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - error finding items", "error", err)
		return err
	}

	validBlueprints := []models.OwnedBlueprint{}
	for _, uniqueName := range req.UniqueNames {
		item, exists := items[uniqueName]
		if !exists {
			logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - item not found, skipping", "uniqueName", uniqueName)
			continue
		}
		if item.ConsumeOnBuild {
			logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - blueprint not reusable, skipping", "uniqueName", uniqueName)
			continue
		}
		validBlueprints = append(validBlueprints, models.OwnedBlueprint{
			UniqueName: uniqueName,
			AddedAt:    time.Now(),
		})
	}

	if len(validBlueprints) == 0 {
		logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - no valid blueprints to add")
		return nil
	}

	// Get existing owned blueprints to filter duplicates
	ownedBP, err := s.ownedBPRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - error fetching owned blueprints", "error", err)
		return err
	}

	existingSet := make(map[string]bool)
	if ownedBP != nil {
		for _, bp := range ownedBP.Blueprints {
			existingSet[bp.UniqueName] = true
		}
	}

	// Filter out already owned blueprints
	newBlueprints := []models.OwnedBlueprint{}
	for _, bp := range validBlueprints {
		if !existingSet[bp.UniqueName] {
			newBlueprints = append(newBlueprints, bp)
		}
	}

	if len(newBlueprints) == 0 {
		logger.Debug(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - all blueprints already owned")
		return nil
	}

	// Create if doesn't exist, then bulk add
	if ownedBP == nil {
		ownedBP = &models.OwnedBlueprints{
			UserID:     userID,
			Blueprints: newBlueprints,
		}
		err = s.ownedBPRepo.Create(ctx, ownedBP)
		if err != nil {
			logger.Error(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - error creating owned blueprints", "error", err)
			return err
		}
	} else {
		err = s.ownedBPRepo.BulkAddBlueprints(ctx, userID, newBlueprints)
		if err != nil {
			logger.Error(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - error bulk adding blueprints", "error", err)
			return err
		}
	}

	logger.Info(ctx, "service: OwnedBlueprintsService.BulkAddBlueprints - blueprints added successfully", "count", len(newBlueprints))
	return nil
}

func (s *OwnedBlueprintsService) ClearAllBlueprints(ctx context.Context, userID string) error {
	logger.Debug(ctx, "service: OwnedBlueprintsService.ClearAllBlueprints called", "userID", userID)

	ownedBP, err := s.ownedBPRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.ClearAllBlueprints - error fetching owned blueprints", "error", err)
		return err
	}

	if ownedBP == nil {
		logger.Debug(ctx, "service: OwnedBlueprintsService.ClearAllBlueprints - no owned blueprints to clear")
		return nil
	}

	err = s.ownedBPRepo.ClearAll(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: OwnedBlueprintsService.ClearAllBlueprints - error clearing blueprints", "error", err)
		return err
	}

	logger.Info(ctx, "service: OwnedBlueprintsService.ClearAllBlueprints - all blueprints cleared successfully")
	return nil
}
