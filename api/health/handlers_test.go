package health

import (
	"net/http"
	"testing"

	"github.com/ingka-group-digital/ocp-go-utils/api/test"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func TestIntegrationHealthHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("(skipped)")
	}

	it := test.NewIntegrationTest(t, test.IntegrationTestWithDatabase{})
	defer func() {
		it.TearDown()
	}()

	healthHandler := NewHealthHandler(it.Db)

	tests := []test.Data{
		{
			Name:       "Ready probe ok",
			Method:     http.MethodGet,
			Handler:    healthHandler.Ready,
			ExpectCode: http.StatusNoContent,
		},
		{
			Name:   "Ready probe unavailable",
			Method: http.MethodGet,
			Handler: func(ctx echo.Context) error {
				// drop database connection
				it.Db, _ = gorm.Open(nil)

				hh := NewHealthHandler(it.Db)
				return hh.Ready(ctx)
			},
			ExpectCode: http.StatusServiceUnavailable,
		},
		{
			Name:           "Live probe ok",
			Method:         http.MethodGet,
			Handler:        healthHandler.Live,
			ExpectCode:     http.StatusOK,
			ExpectResponse: "live-probe-ok",
		},
		{
			Name:   "Live probe unavailable",
			Method: http.MethodGet,
			Handler: func(ctx echo.Context) error {
				// drop database connection
				it.Db, _ = gorm.Open(nil)

				hh := NewHealthHandler(it.Db)
				return hh.Live(ctx)
			},
			ExpectCode:     http.StatusServiceUnavailable,
			ExpectResponse: "live-probe-unavailable",
		},
	}

	test.Assert(it, tests)
}
