package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/internal/services"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"github.com/graytonio/warframe-wishlist/pkg/response"
)

type ItemHandler struct {
	itemService services.ItemServiceInterface
}

func NewItemHandler(itemService services.ItemServiceInterface) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	params := models.SearchParams{
		Query:    query.Get("q"),
		Category: query.Get("category"),
		Limit:    limit,
		Offset:   offset,
	}

	logger.Debug(ctx, "handler: Search called", "query", params.Query, "category", params.Category, "limit", params.Limit, "offset", params.Offset)

	items, err := h.itemService.Search(ctx, params)
	if err != nil {
		logger.Error(ctx, "handler: Search - failed to search items", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to search items")
		return
	}

	logger.Info(ctx, "handler: Search - success", "resultCount", len(items))
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (h *ItemHandler) GetByUniqueName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Use wildcard param to capture full path including slashes (e.g., /Lotus/Types/Items/...)
	uniqueName := chi.URLParam(r, "*")
	if uniqueName == "" {
		logger.Warn(ctx, "handler: GetByUniqueName - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	// Add leading slash to the uniqueName
	uniqueName = "/" + uniqueName

	logger.Debug(ctx, "handler: GetByUniqueName called", "uniqueName", uniqueName)

	item, err := h.itemService.GetByUniqueName(ctx, uniqueName)
	if err != nil {
		logger.Error(ctx, "handler: GetByUniqueName - failed to get item", "error", err, "uniqueName", uniqueName)
		response.Error(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	if item == nil {
		logger.Warn(ctx, "handler: GetByUniqueName - item not found", "uniqueName", uniqueName)
		response.Error(w, http.StatusNotFound, "item not found")
		return
	}

	logger.Info(ctx, "handler: GetByUniqueName - success", "uniqueName", uniqueName, "itemName", item.Name)
	response.JSON(w, http.StatusOK, item)
}

func (h *ItemHandler) SearchReusableBlueprints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	q := query.Get("q")

	logger.Debug(ctx, "handler: SearchReusableBlueprints called", "query", q, "limit", limit)

	items, err := h.itemService.SearchReusableBlueprints(ctx, q, limit)
	if err != nil {
		logger.Error(ctx, "handler: SearchReusableBlueprints - failed to search reusable blueprints", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to search reusable blueprints")
		return
	}

	logger.Info(ctx, "handler: SearchReusableBlueprints - success", "resultCount", len(items))
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}
