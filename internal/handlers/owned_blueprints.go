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

type OwnedBlueprintsHandler struct {
	ownedBPService services.OwnedBlueprintsServiceInterface
}

func NewOwnedBlueprintsHandler(ownedBPService services.OwnedBlueprintsServiceInterface) *OwnedBlueprintsHandler {
	return &OwnedBlueprintsHandler{
		ownedBPService: ownedBPService,
	}
}

func (h *OwnedBlueprintsHandler) GetOwnedBlueprints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: GetOwnedBlueprints called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: GetOwnedBlueprints - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	logger.Debug(ctx, "handler: GetOwnedBlueprints - fetching owned blueprints", "userID", userID)
	ownedBP, err := h.ownedBPService.GetOwnedBlueprints(ctx, userID)
	if err != nil {
		logger.Error(ctx, "handler: GetOwnedBlueprints - failed to get owned blueprints", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to get owned blueprints")
		return
	}

	blueprintCount := 0
	if ownedBP != nil {
		blueprintCount = len(ownedBP.Blueprints)
	}
	logger.Info(ctx, "handler: GetOwnedBlueprints - success", "blueprintCount", blueprintCount)
	response.JSON(w, http.StatusOK, ownedBP)
}

func (h *OwnedBlueprintsHandler) AddBlueprint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: AddBlueprint called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: AddBlueprint - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req models.AddBlueprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(ctx, "handler: AddBlueprint - invalid request body", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UniqueName == "" {
		logger.Warn(ctx, "handler: AddBlueprint - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	logger.Debug(ctx, "handler: AddBlueprint - adding blueprint", "uniqueName", req.UniqueName)
	err := h.ownedBPService.AddBlueprint(ctx, userID, req)
	if err != nil {
		if errors.Is(err, services.ErrBlueprintNotFound) {
			logger.Warn(ctx, "handler: AddBlueprint - blueprint not found", "uniqueName", req.UniqueName)
			response.Error(w, http.StatusNotFound, "blueprint not found")
			return
		}
		if errors.Is(err, services.ErrBlueprintNotReusable) {
			logger.Warn(ctx, "handler: AddBlueprint - blueprint not reusable", "uniqueName", req.UniqueName)
			response.Error(w, http.StatusBadRequest, "blueprint is not reusable (consumeOnBuild is true)")
			return
		}
		if errors.Is(err, services.ErrBlueprintAlreadyOwned) {
			logger.Warn(ctx, "handler: AddBlueprint - blueprint already owned", "uniqueName", req.UniqueName)
			response.Error(w, http.StatusConflict, "blueprint already owned")
			return
		}
		logger.Error(ctx, "handler: AddBlueprint - failed to add blueprint", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to add blueprint")
		return
	}

	logger.Info(ctx, "handler: AddBlueprint - success", "uniqueName", req.UniqueName)
	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "blueprint added",
	})
}

func (h *OwnedBlueprintsHandler) RemoveBlueprint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: RemoveBlueprint called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: RemoveBlueprint - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Use wildcard param to capture full path including slashes (e.g., /Lotus/Types/Items/...)
	uniqueName := chi.URLParam(r, "*")
	if uniqueName == "" {
		logger.Warn(ctx, "handler: RemoveBlueprint - uniqueName is required")
		response.Error(w, http.StatusBadRequest, "uniqueName is required")
		return
	}

	// Add leading slash to the uniqueName
	uniqueName = "/" + uniqueName

	logger.Debug(ctx, "handler: RemoveBlueprint - removing blueprint", "uniqueName", uniqueName)
	err := h.ownedBPService.RemoveBlueprint(ctx, userID, uniqueName)
	if err != nil {
		if errors.Is(err, services.ErrBlueprintNotOwned) {
			logger.Warn(ctx, "handler: RemoveBlueprint - blueprint not owned", "uniqueName", uniqueName)
			response.Error(w, http.StatusNotFound, "blueprint not owned")
			return
		}
		logger.Error(ctx, "handler: RemoveBlueprint - failed to remove blueprint", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to remove blueprint")
		return
	}

	logger.Info(ctx, "handler: RemoveBlueprint - success", "uniqueName", uniqueName)
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "blueprint removed",
	})
}

func (h *OwnedBlueprintsHandler) BulkAddBlueprints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: BulkAddBlueprints called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: BulkAddBlueprints - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req models.BulkAddBlueprintsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(ctx, "handler: BulkAddBlueprints - invalid request body", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	logger.Debug(ctx, "handler: BulkAddBlueprints - bulk adding blueprints", "count", len(req.UniqueNames))
	err := h.ownedBPService.BulkAddBlueprints(ctx, userID, req)
	if err != nil {
		logger.Error(ctx, "handler: BulkAddBlueprints - failed to bulk add blueprints", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to bulk add blueprints")
		return
	}

	logger.Info(ctx, "handler: BulkAddBlueprints - success", "count", len(req.UniqueNames))
	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "blueprints added",
	})
}

func (h *OwnedBlueprintsHandler) ClearAllBlueprints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: ClearAllBlueprints called")

	userID := middleware.GetUserID(ctx)
	if userID == "" {
		logger.Warn(ctx, "handler: ClearAllBlueprints - user not authenticated")
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	logger.Debug(ctx, "handler: ClearAllBlueprints - clearing all blueprints")
	err := h.ownedBPService.ClearAllBlueprints(ctx, userID)
	if err != nil {
		logger.Error(ctx, "handler: ClearAllBlueprints - failed to clear blueprints", "error", err)
		response.Error(w, http.StatusInternalServerError, "failed to clear blueprints")
		return
	}

	logger.Info(ctx, "handler: ClearAllBlueprints - success")
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "all blueprints cleared",
	})
}
