// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package echozap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	trace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ingka-group/fastecho/fctx"
)

func TestZapLoggerMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusBadRequest, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLoggerMiddleware(logger)(h)(c)

	assert.Nil(t, err)

	logFields := logs.AllUntimed()[0].ContextMap()

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(400), logFields["status"])
	assert.IsType(t, float64(0), logFields["latency_ms"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLoggerMiddlewareWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusBadRequest, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLoggerMiddlewareWithConfig(logger, ZapLoggerMiddlewareConfig{
		Skipper: func(ctx echo.Context) bool {
			return strings.Contains(ctx.Request().URL.Path, "/something")
		},
	})(h)(c)

	assert.Nil(t, err)

	assert.Equal(t, 0, logs.Len())
}

func TestAccessLog_5xxCarriesTraceIDAndError(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x01},
		TraceFlags: trace.FlagsSampled,
	})

	e := echo.New()
	// Seed a span context, as otelecho would, so Fields(ctx) yields trace_id.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := trace.ContextWithSpanContext(c.Request().Context(), sc)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	e.Use(ZapLoggerMiddlewareWithConfig(zap.New(core), ZapLoggerMiddlewareConfig{}))
	// Return an error so the access log captures it (zap.Error).
	e.GET("/x", func(c echo.Context) error { return echo.NewHTTPError(500, "boom") })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil))

	require.Equal(t, 1, logs.Len(), "5xx still logged (regression guard)")
	got := logs.All()[0]
	assert.Equal(t, "Server error", got.Message)
	assert.Equal(t, sc.TraceID().String(), got.ContextMap()["trace_id"], "5xx line carries trace_id")
	assert.Contains(t, got.ContextMap(), "error", "5xx line carries the error string")
}

func TestAccessLog_4xxCarriesRequestID(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := fctx.WithRequestID(c.Request().Context(), "req-1")
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	e.Use(ZapLoggerMiddlewareWithConfig(zap.New(core), ZapLoggerMiddlewareConfig{}))
	e.GET("/x", func(c echo.Context) error { return c.String(400, "bad") })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil))

	require.Equal(t, 1, logs.Len())
	assert.Equal(t, "Client error", logs.All()[0].Message)
	assert.Equal(t, "req-1", logs.All()[0].ContextMap()["request_id"])
}

func TestAccessLog_UsesResponseRequestIDWhenContextHasNone(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	e := echo.New()
	e.Use(ZapLoggerMiddlewareWithConfig(zap.New(core), ZapLoggerMiddlewareConfig{}))
	e.GET("/x", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderXRequestID, "header-req")
		return c.String(http.StatusBadRequest, "bad")
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil))

	require.Equal(t, 1, logs.Len())
	assert.Equal(t, "header-req", logs.All()[0].ContextMap()["request_id"])
}

func TestAccessLog_2xxLoggedAtDebug(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	e := echo.New()
	e.Use(ZapLoggerMiddlewareWithConfig(zap.New(core), ZapLoggerMiddlewareConfig{}))
	e.GET("/x", func(c echo.Context) error { return c.String(200, "ok") })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil))

	require.Equal(t, 1, logs.Len(), "2xx logged")
	got := logs.All()[0]
	assert.Equal(t, zapcore.DebugLevel, got.Level, "2xx logged at Debug")
	assert.Equal(t, "Success", got.Message)
	assert.Equal(t, "/x", got.ContextMap()["path"], "path is the route template")
}
