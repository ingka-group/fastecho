package health

import (
	"net/http"
	"testing"

	"github.com/ingka-group-digital/ocp-go-utils/api/test"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func TestIntegrationHandler_Ready(t *testing.T) {
	if testing.Short() {
		t.Skip("(skipped)")
	}

	it := test.NewIntegrationTest(t, test.IntegrationTestWithPostgres{})
	defer func() {
		it.TearDown()
	}()

	healthHandler := NewHandler(it.Db)

	tests := []test.Data{
		{
			Name:       "ok: Ready probe",
			Method:     http.MethodGet,
			Handler:    healthHandler.Ready,
			ExpectCode: http.StatusOK,
		},
		{
			Name:       "ok: No database",
			Method:     http.MethodGet,
			Handler:    NewHandler(nil).Ready,
			ExpectCode: http.StatusOK,
		},
		{
			Name:   "fail: Ready probe unavailable",
			Method: http.MethodGet,
			Handler: func(ctx echo.Context) error {
				// drop database connection
				it.Db, _ = gorm.Open(nil)

				hh := NewHandler(it.Db)
				return hh.Ready(ctx)
			},
			ExpectCode: http.StatusServiceUnavailable,
		},
	}

	test.AssertAll(it, tests)
}

func TestIntegrationHandler_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("(skipped)")
	}

	it := test.NewIntegrationTest(t, test.IntegrationTestWithPostgres{})
	defer func() {
		it.TearDown()
	}()

	healthHandler := NewHandler(it.Db)

	tests := []test.Data{
		{
			Name:           "ok: Live probe",
			Method:         http.MethodGet,
			Handler:        healthHandler.Live,
			ExpectCode:     http.StatusOK,
			ExpectResponse: "live-probe-ok",
		},
		{
			Name:           "ok: No database",
			Method:         http.MethodGet,
			Handler:        NewHandler(nil).Live,
			ExpectCode:     http.StatusOK,
			ExpectResponse: "live-probe-ok",
		},
		{
			Name:   "fail: Live probe unavailable",
			Method: http.MethodGet,
			Handler: func(ctx echo.Context) error {
				// drop database connection
				it.Db, _ = gorm.Open(nil)

				hh := NewHandler(it.Db)
				return hh.Live(ctx)
			},
			ExpectCode:     http.StatusServiceUnavailable,
			ExpectResponse: "live-probe-unavailable",
		},
	}

	test.AssertAll(it, tests)
}
