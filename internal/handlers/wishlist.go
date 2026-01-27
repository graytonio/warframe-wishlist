package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/middleware"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/services"
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
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	wishlist, err := h.wishlistService.GetWishlist(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get wishlist")
		return
	}

	response.JSON(w, http.StatusOK, wishlist)
}

func (h *WishlistHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req models.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UniqueName == "" {
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	err := h.wishlistService.AddItem(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			response.Error(w, http.StatusNotFound, "item not found")
			return
		}
		if errors.Is(err, services.ErrItemAlreadyInWishlist) {
			response.Error(w, http.StatusConflict, "item already in wishlist")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to add item to wishlist")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "item added to wishlist",
	})
}

func (h *WishlistHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	uniqueName := chi.URLParam(r, "uniqueName")
	if uniqueName == "" {
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	err := h.wishlistService.RemoveItem(r.Context(), userID, uniqueName)
	if err != nil {
		if errors.Is(err, services.ErrItemNotInWishlist) {
			response.Error(w, http.StatusNotFound, "item not in wishlist")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to remove item from wishlist")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "item removed from wishlist",
	})
}

func (h *WishlistHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	uniqueName := chi.URLParam(r, "uniqueName")
	if uniqueName == "" {
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	var req models.UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.wishlistService.UpdateQuantity(r.Context(), userID, uniqueName, req.Quantity)
	if err != nil {
		if errors.Is(err, services.ErrItemNotInWishlist) {
			response.Error(w, http.StatusNotFound, "item not in wishlist")
			return
		}
		if errors.Is(err, services.ErrInvalidQuantity) {
			response.Error(w, http.StatusBadRequest, "quantity must be greater than 0")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update quantity")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "quantity updated",
	})
}

func (h *WishlistHandler) GetMaterials(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	materials, err := h.materialResolver.GetMaterials(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get materials")
		return
	}

	response.JSON(w, http.StatusOK, materials)
}
