package date

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestISODate_UnmarshalJSON(t *testing.T) {
	type result struct {
		Created ISODate `json:"created"`
	}

	tests := []struct {
		name          string
		given         string
		expectCreated time.Time
	}{
		{
			name:          "ok: date given",
			given:         "{\"created\": \"2021-01-10\"}",
			expectCreated: time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := result{}
			err := json.Unmarshal([]byte(tt.given), &data)
			if err != nil {
				t.Fail()
			}

			require.Equal(t, tt.expectCreated, data.Created.Time)
		})
	}
}

func TestISODate_UnmarshalParam(t *testing.T) {
	type param struct {
		Created ISODate `query:"created"`
	}

	tests := []struct {
		name          string
		given         string
		expectCreated time.Time
		expectErr     bool
	}{
		{
			name:          "ok: date given",
			given:         "2021-01-10",
			expectCreated: time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "ok: invalid date given",
			given:     "2021-01-10T10:15",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/something", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set query parameter to the context
			q := c.QueryParams()
			q.Add("created", tt.given) // query param name should match the name used in struct param

			// Create a handler on the fly
			h := func(c echo.Context) error {
				p := new(param)
				if err := c.Bind(p); err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, err.Error())
				}

				require.Equal(t, tt.expectCreated, p.Created.Time)
				return c.NoContent(http.StatusOK)
			}

			// Call the handler with the context created before
			err := h(c)
			if err != nil {
				require.True(t, tt.expectErr)
			}
		})
	}
}
