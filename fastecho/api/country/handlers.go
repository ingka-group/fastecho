package country

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CountryHandler defines the http router implementation for country endpoints.
type CountryHandler struct{}

// NewCountryHandler creates a new CountryHandler.
func NewCountryHandler() *CountryHandler {
	return &CountryHandler{}
}

// GetCountries endpoint returns the list of known countries.
//
// @Summary Get countries
// @Description Returns the list of known countries
// @Tags country
// @ID get-countries
// @Produce json
// @Success 200 {object} Countries "OK"
// @Router /v1/countries [get]
func (h *CountryHandler) GetCountries(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, SortedCountries())
}
