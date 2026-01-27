package services

import (
	"context"
	"errors"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/repository"
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
	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if wishlist == nil {
		wishlist = &models.Wishlist{
			UserID:    userID,
			Items:     []models.WishlistItem{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return wishlist, nil
}

func (s *WishlistService) AddItem(ctx context.Context, userID string, req models.AddItemRequest) error {
	item, err := s.itemRepo.FindByUniqueName(ctx, req.UniqueName)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if wishlist == nil {
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
		return s.wishlistRepo.Create(ctx, wishlist)
	}

	for _, wi := range wishlist.Items {
		if wi.UniqueName == req.UniqueName {
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

	return s.wishlistRepo.AddItem(ctx, userID, newItem)
}

func (s *WishlistService) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if wishlist == nil {
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
		return ErrItemNotInWishlist
	}

	return s.wishlistRepo.RemoveItem(ctx, userID, uniqueName)
}

func (s *WishlistService) UpdateQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	wishlist, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if wishlist == nil {
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
		return ErrItemNotInWishlist
	}

	return s.wishlistRepo.UpdateItemQuantity(ctx, userID, uniqueName, quantity)
}
