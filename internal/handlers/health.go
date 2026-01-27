package handlers

import (
	"net/http"

	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"github.com/graytonio/warframe-wishlist/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(ctx, "handler: Health called")
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
