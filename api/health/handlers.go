package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// HealthHandler defines the http router implementation for health endpoints.
type HealthHandler struct {
	DB *gorm.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		DB: db,
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
// @Router /v1/health/ready [get]
func (h *HealthHandler) Ready(ctx echo.Context) error {
	if CheckDatabase(h.DB) != nil {
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
// @Router /v1/health/live [get]
func (h *HealthHandler) Live(ctx echo.Context) error {
	if CheckDatabase(h.DB) != nil {
		return ctx.JSON(http.StatusServiceUnavailable, ServiceHealth{
			ServiceStatus: StatusUnhealthy,
			Description:   DescriptionDatabaseIsDown,
		})
	}

	return ctx.JSON(http.StatusOK, ServiceHealth{
		ServiceStatus: StatusHealthy,
		Description:   DescriptionHealthy,
	})
}
