package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/services"
	"github.com/graytonio/warframe-wishlist/pkg/response"
)

type ItemHandler struct {
	itemService services.ItemServiceInterface
}

func NewItemHandler(itemService services.ItemServiceInterface) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	params := models.SearchParams{
		Query:    query.Get("q"),
		Category: query.Get("category"),
		Limit:    limit,
		Offset:   offset,
	}

	items, err := h.itemService.Search(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to search items")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (h *ItemHandler) GetByUniqueName(w http.ResponseWriter, r *http.Request) {
	uniqueName := chi.URLParam(r, "uniqueName")
	if uniqueName == "" {
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	item, err := h.itemService.GetByUniqueName(r.Context(), uniqueName)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	if item == nil {
		response.Error(w, http.StatusNotFound, "item not found")
		return
	}

	response.JSON(w, http.StatusOK, item)
}
