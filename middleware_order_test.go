// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fastecho

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/router" // Config.Routes is func(*echo.Echo, *router.Router) error
)

func TestRecover_PanicReturns500(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	fe, err := Initialize(&Config{
		Routes: func(e *echo.Echo, r *router.Router) error {
			e.GET("/boom", func(c echo.Context) error { panic("kaboom") })
			return nil
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = fe.Shutdown(context.Background()) })

	rec := httptest.NewRecorder()
	fe.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	// recover turns the panic into a 500 instead of crashing the server.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServerSpan_CarriesRequestID(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	e := echo.New()
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{Generator: fctx.NewRequestID}))
	e.Use(otelecho.Middleware("test", otelecho.WithTracerProvider(tp)))
	e.Use(fctx.Middleware(zap.NewNop(), tp.Tracer("t")))
	e.GET("/x", func(c echo.Context) error { return c.String(200, "ok") })

	r := httptest.NewRecorder()
	e.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/x", nil))

	spans := rec.Ended()
	require.Len(t, spans, 1)
	var reqID string
	for _, kv := range spans[0].Attributes() {
		if string(kv.Key) == "fastecho.request_id" {
			reqID = kv.Value.AsString()
		}
	}
	assert.NotEmpty(t, reqID, "server span carries fastecho.request_id")
}
