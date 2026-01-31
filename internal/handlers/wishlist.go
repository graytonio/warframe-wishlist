package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/middleware"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/services"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"github.com/graytonio/warframe-wishlist/pkg/response"
)

type WishlistHandler struct {
	wishlistService  services.WishlistServiceInterface
	materialResolver services.MaterialResolverInterface
}

func NewWishlistHandler(wishlistService services.WishlistServiceInterface, materialResolver services.MaterialResolverInterface) *WishlistHandler {
	return &WishlistHandler{
		wishlistService:  wishlistService,
		materialResolver: materialResolver,
	}
}

func (h *WishlistHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: GetWishlist called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: GetWishlist - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	logger.Debug(ctx, "handler: GetWishlist - fetching wishlist", "userID", userID)
	wishlist, err := h.wishlistService.GetWishlist(ctx, userID)
	if err != nil {
		logger.Error(ctx, "handler: GetWishlist - failed to get wishlist", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to get wishlist")
		return
	}

	itemCount := 0
	if wishlist != nil {
		itemCount = len(wishlist.Items)
	}
	logger.Info(ctx, "handler: GetWishlist - success", "itemCount", itemCount)
	response.JSON(w, http.StatusOK, wishlist)
}

func (h *WishlistHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: AddItem called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: AddItem - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req models.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(ctx, "handler: AddItem - invalid request body", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UniqueName == "" {
		logger.Warn(ctx, "handler: AddItem - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	logger.Debug(ctx, "handler: AddItem - adding item to wishlist", "uniqueName", req.UniqueName, "quantity", req.Quantity)
	err := h.wishlistService.AddItem(ctx, userID, req)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			logger.Warn(ctx, "handler: AddItem - item not found", "uniqueName", req.UniqueName)
			response.Error(w, http.StatusNotFound, "item not found")
			return
		}
		if errors.Is(err, services.ErrItemAlreadyInWishlist) {
			logger.Warn(ctx, "handler: AddItem - item already in wishlist", "uniqueName", req.UniqueName)
			response.Error(w, http.StatusConflict, "item already in wishlist")
			return
		}
		logger.Error(ctx, "handler: AddItem - failed to add item to wishlist", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to add item to wishlist")
		return
	}

	logger.Info(ctx, "handler: AddItem - success", "uniqueName", req.UniqueName)
	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "item added to wishlist",
	})
}

func (h *WishlistHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: RemoveItem called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: RemoveItem - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Use wildcard param to capture full path including slashes (e.g., /Lotus/Types/Items/...)
	uniqueName := chi.URLParam(r, "*")
	if uniqueName == "" {
		logger.Warn(ctx, "handler: RemoveItem - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	// Add leading slash to the uniqueName
	uniqueName = "/" + uniqueName

	logger.Debug(ctx, "handler: RemoveItem - removing item from wishlist", "uniqueName", uniqueName)
	err := h.wishlistService.RemoveItem(ctx, userID, uniqueName)
	if err != nil {
		if errors.Is(err, services.ErrItemNotInWishlist) {
			logger.Warn(ctx, "handler: RemoveItem - item not in wishlist", "uniqueName", uniqueName)
			response.Error(w, http.StatusNotFound, "item not in wishlist")
			return
		}
		logger.Error(ctx, "handler: RemoveItem - failed to remove item from wishlist", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to remove item from wishlist")
		return
	}

	logger.Info(ctx, "handler: RemoveItem - success", "uniqueName", uniqueName)
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "item removed from wishlist",
	})
}

func (h *WishlistHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: UpdateQuantity called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: UpdateQuantity - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Use wildcard param to capture full path including slashes (e.g., /Lotus/Types/Items/...)
	uniqueName := chi.URLParam(r, "*")
	if uniqueName == "" {
		logger.Warn(ctx, "handler: UpdateQuantity - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	var req models.UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(ctx, "handler: UpdateQuantity - invalid request body", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	logger.Debug(ctx, "handler: UpdateQuantity - updating quantity", "uniqueName", uniqueName, "quantity", req.Quantity)
	err := h.wishlistService.UpdateQuantity(ctx, userID, uniqueName, req.Quantity)
	if err != nil {
		if errors.Is(err, services.ErrItemNotInWishlist) {
			logger.Warn(ctx, "handler: UpdateQuantity - item not in wishlist", "uniqueName", uniqueName)
			response.Error(w, http.StatusNotFound, "item not in wishlist")
			return
		}
		if errors.Is(err, services.ErrInvalidQuantity) {
			logger.Warn(ctx, "handler: UpdateQuantity - invalid quantity", "quantity", req.Quantity)
			response.Error(w, http.StatusBadRequest, "quantity must be greater than 0")
			return
		}
		logger.Error(ctx, "handler: UpdateQuantity - failed to update quantity", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to update quantity")
		return
	}

	logger.Info(ctx, "handler: UpdateQuantity - success", "uniqueName", uniqueName, "quantity", req.Quantity)
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "quantity updated",
	})
}

func (h *WishlistHandler) GetMaterials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: GetMaterials called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: GetMaterials - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	logger.Debug(ctx, "handler: GetMaterials - resolving materials")
	materials, err := h.materialResolver.GetMaterials(ctx, userID)
	if err != nil {
		logger.Error(ctx, "handler: GetMaterials - failed to get materials", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to get materials")
		return
	}

	materialCount := 0
	if materials != nil {
		materialCount = len(materials.Materials)
	}
	logger.Info(ctx, "handler: GetMaterials - success", "materialCount", materialCount, "totalCredits", materials.TotalCredits)
	response.JSON(w, http.StatusOK, materials)
}
