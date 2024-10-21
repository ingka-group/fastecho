package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines the http router implementation for health endpoints.
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new Handler for health endpoints.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// Ready performs readiness check.
//
// @Summary Ready healthcheck
// @Description Performs readiness check
// @Tags health
// @ID health-ready
// @Success 200 "OK"
// @Failure 503 {object} ServiceHealth "Service Unavailable"
// @Router /health/ready [get]
func (h *Handler) Ready(ctx echo.Context) error {
	if checkDatabase(h.db) != nil {
		return ctx.NoContent(http.StatusServiceUnavailable)
	}

	return ctx.NoContent(http.StatusOK)
}

// Live performs a live check.
//
// @Summary Live healthcheck
// @Description Performs a live check
// @Tags health
// @ID health-live
// @Produce json
// @Success 200 {object} ServiceHealth "OK"
// @Failure 503 {object} ServiceHealth "Service Unavailable"
// @Router /health/live [get]
func (h *Handler) Live(ctx echo.Context) error {
	if checkDatabase(h.db) != nil {
		return ctx.JSON(http.StatusServiceUnavailable, ServiceHealth{
			ServiceStatus: statusUnhealthy,
			Description:   descriptionDatabaseIsDown,
		})
	}

	return ctx.JSON(http.StatusOK, ServiceHealth{
		ServiceStatus: statusHealthy,
		Description:   descriptionHealthy,
	})
}
