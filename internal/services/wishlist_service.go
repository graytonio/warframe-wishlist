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
	ErrItemAlreadyInWishlist = errors.New("item already in wishlist")
	ErrItemNotFound          = errors.New("item not found")
	ErrItemNotInWishlist     = errors.New("item not in wishlist")
	ErrInvalidQuantity       = errors.New("quantity must be greater than 0")
)

type WishlistService struct {
	wishlistRepo repository.WishlistRepositoryInterface
	itemRepo     repository.ItemRepositoryInterface
}

func NewWishlistService(wishlistRepo repository.WishlistRepositoryInterface, itemRepo repository.ItemRepositoryInterface) *WishlistService {
	return &WishlistService{
		wishlistRepo: wishlistRepo,
		itemRepo:     itemRepo,
	}
}

func (s *WishlistService) GetWishlist(ctx context.Context, userID string) (*models.Wishlist, error) {
	logger.Debug(ctx, "service: WishlistService.GetWishlist called", "userID", userID)

	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.GetWishlist - repository error", "error", err)
		return nil, err
	}

	if wishlist == nil {
		logger.Debug(ctx, "service: WishlistService.GetWishlist - creating empty wishlist for new user")
		wishlist = &models.Wishlist{
			UserID:    userID,
			Items:     []models.WishlistItem{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	logger.Debug(ctx, "service: WishlistService.GetWishlist - completed", "itemCount", len(wishlist.Items))
	return wishlist, nil
}

func (s *WishlistService) AddItem(ctx context.Context, userID string, req models.AddItemRequest) error {
	logger.Debug(ctx, "service: WishlistService.AddItem called", "userID", userID, "uniqueName", req.UniqueName, "quantity", req.Quantity)

	logger.Debug(ctx, "service: WishlistService.AddItem - validating item exists")
	item, err := s.itemRepo.FindByUniqueName(ctx, req.UniqueName)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.AddItem - error finding item", "error", err)
		return err
	}
	if item == nil {
		logger.Warn(ctx, "service: WishlistService.AddItem - item not found", "uniqueName", req.UniqueName)
		return ErrItemNotFound
	}

	logger.Debug(ctx, "service: WishlistService.AddItem - fetching user wishlist")
	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.AddItem - error fetching wishlist", "error", err)
		return err
	}

	if wishlist == nil {
		logger.Debug(ctx, "service: WishlistService.AddItem - creating new wishlist for user")
		quantity := req.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		wishlist = &models.Wishlist{
			UserID: userID,
			Items: []models.WishlistItem{
				{
					UniqueName: req.UniqueName,
					Quantity:   quantity,
					AddedAt:    time.Now(),
				},
			},
		}
		err = s.wishlistRepo.Create(ctx, wishlist)
		if err != nil {
			logger.Error(ctx, "service: WishlistService.AddItem - error creating wishlist", "error", err)
			return err
		}
		logger.Info(ctx, "service: WishlistService.AddItem - created new wishlist with item", "uniqueName", req.UniqueName)
		return nil
	}

	for _, wi := range wishlist.Items {
		if wi.UniqueName == req.UniqueName {
			logger.Warn(ctx, "service: WishlistService.AddItem - item already in wishlist", "uniqueName", req.UniqueName)
			return ErrItemAlreadyInWishlist
		}
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	newItem := models.WishlistItem{
		UniqueName: req.UniqueName,
		Quantity:   quantity,
		AddedAt:    time.Now(),
	}

	err = s.wishlistRepo.AddItem(ctx, userID, newItem)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.AddItem - error adding item to wishlist", "error", err)
		return err
	}
	logger.Info(ctx, "service: WishlistService.AddItem - item added successfully", "uniqueName", req.UniqueName, "quantity", quantity)
	return nil
}

func (s *WishlistService) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	logger.Debug(ctx, "service: WishlistService.RemoveItem called", "userID", userID, "uniqueName", uniqueName)

	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.RemoveItem - error fetching wishlist", "error", err)
		return err
	}

	if wishlist == nil {
		logger.Warn(ctx, "service: WishlistService.RemoveItem - wishlist not found for user")
		return ErrItemNotInWishlist
	}

	found := false
	for _, wi := range wishlist.Items {
		if wi.UniqueName == uniqueName {
			found = true
			break
		}
	}

	if !found {
		logger.Warn(ctx, "service: WishlistService.RemoveItem - item not in wishlist", "uniqueName", uniqueName)
		return ErrItemNotInWishlist
	}

	err = s.wishlistRepo.RemoveItem(ctx, userID, uniqueName)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.RemoveItem - error removing item", "error", err)
		return err
	}
	logger.Info(ctx, "service: WishlistService.RemoveItem - item removed successfully", "uniqueName", uniqueName)
	return nil
}

func (s *WishlistService) UpdateQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	logger.Debug(ctx, "service: WishlistService.UpdateQuantity called", "userID", userID, "uniqueName", uniqueName, "quantity", quantity)

	if quantity <= 0 {
		logger.Warn(ctx, "service: WishlistService.UpdateQuantity - invalid quantity", "quantity", quantity)
		return ErrInvalidQuantity
	}

	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.UpdateQuantity - error fetching wishlist", "error", err)
		return err
	}

	if wishlist == nil {
		logger.Warn(ctx, "service: WishlistService.UpdateQuantity - wishlist not found for user")
		return ErrItemNotInWishlist
	}

	found := false
	for _, wi := range wishlist.Items {
		if wi.UniqueName == uniqueName {
			found = true
			break
		}
	}

	if !found {
		logger.Warn(ctx, "service: WishlistService.UpdateQuantity - item not in wishlist", "uniqueName", uniqueName)
		return ErrItemNotInWishlist
	}

	err = s.wishlistRepo.UpdateItemQuantity(ctx, userID, uniqueName, quantity)
	if err != nil {
		logger.Error(ctx, "service: WishlistService.UpdateQuantity - error updating quantity", "error", err)
		return err
	}
	logger.Info(ctx, "service: WishlistService.UpdateQuantity - quantity updated successfully", "uniqueName", uniqueName, "quantity", quantity)
	return nil
}
