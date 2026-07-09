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

package fctx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ingka-group/fastecho/fctx"
)

type testFixture struct {
	ec     echo.Context
	logger *zap.Logger
	tracer trace.Tracer
}

func newFixture(t *testing.T) testFixture {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return testFixture{
		ec:     e.NewContext(req, rec),
		logger: zaptest.NewLogger(t),
		tracer: noop.NewTracerProvider().Tracer("test"),
	}
}

func TestMiddleware_InjectsLoggerIntoContext(t *testing.T) {
	f := newFixture(t)

	var gotLogger bool
	handler := func(ec echo.Context) error {
		gotLogger = fctx.Logger(fctx.From(ec)) != nil
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)
	assert.True(t, gotLogger)
}

func TestMiddleware_InjectsTracerIntoContext(t *testing.T) {
	f := newFixture(t)

	var gotTracer bool
	handler := func(ec echo.Context) error {
		gotTracer = fctx.Tracer(fctx.From(ec)) != nil
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)
	assert.True(t, gotTracer)
}

func TestMiddleware_NilTracer_DoesNotPanic(t *testing.T) {
	f := newFixture(t)

	var gotTracer bool
	handler := func(ec echo.Context) error {
		gotTracer = fctx.Tracer(fctx.From(ec)) != nil
		return nil
	}

	err := fctx.Middleware(f.logger, nil)(handler)(f.ec)
	require.NoError(t, err)
	assert.True(t, gotTracer, "Tracer() should return noop fallback when nil")
}

func TestMiddleware_RequestIDFromHeader(t *testing.T) {
	f := newFixture(t)
	f.ec.Response().Header().Set(echo.HeaderXRequestID, "test-req-id")

	var gotID string
	handler := func(ec echo.Context) error {
		gotID = fctx.RequestID(fctx.From(ec))
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)
	assert.Equal(t, "test-req-id", gotID)
}

func TestMiddleware_AbsentRequestID_ReturnsEmpty(t *testing.T) {
	f := newFixture(t)

	var gotID string
	handler := func(ec echo.Context) error {
		gotID = fctx.RequestID(fctx.From(ec))
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)
	assert.Equal(t, "", gotID)
}

func TestMiddleware_LoggerEnrichment_NoTraceFieldsWithoutSpan(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	f := newFixture(t)
	f.logger = zap.New(core)

	handler := func(ec echo.Context) error {
		fctx.Logger(fctx.From(ec)).Info("test-log")
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)

	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.NotContains(t, fields, "trace_id", "no trace_id when span is inactive")
	assert.NotContains(t, fields, "span_id", "no span_id when span is inactive")
}

func TestMiddleware_LoggerEnrichment_IncludesRequestID(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	f := newFixture(t)
	f.logger = zap.New(core)
	f.ec.Response().Header().Set(echo.HeaderXRequestID, "enriched-req-id")

	handler := func(ec echo.Context) error {
		fctx.Logger(fctx.From(ec)).Info("test-log")
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
	require.NoError(t, err)

	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.Equal(t, "enriched-req-id", fields["request_id"])
}

func TestMiddleware_ContextSurvivesSubsequentMiddleware(t *testing.T) {
	f := newFixture(t)

	passThroughMW := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ec echo.Context) error { return next(ec) }
	}

	var gotLogger bool
	handler := func(ec echo.Context) error {
		gotLogger = fctx.Logger(fctx.From(ec)) != nil
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(passThroughMW(handler))(f.ec)
	require.NoError(t, err)
	assert.True(t, gotLogger)
}

func TestMiddleware_ContextSurvivesRequestClone(t *testing.T) {
	f := newFixture(t)

	cloningMW := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ec echo.Context) error {
			ec.SetRequest(ec.Request().Clone(ec.Request().Context()))
			return next(ec)
		}
	}

	var gotLogger bool
	handler := func(ec echo.Context) error {
		gotLogger = fctx.Logger(fctx.From(ec)) != nil
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(cloningMW(handler))(f.ec)
	require.NoError(t, err)
	assert.True(t, gotLogger, "Clone preserves context, values survive")
}

func TestMiddleware_ContextLostWhenReplacedWithBackground(t *testing.T) {
	f := newFixture(t)

	destroyingMW := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ec echo.Context) error {
			ec.SetRequest(ec.Request().WithContext(context.Background()))
			return next(ec)
		}
	}

	var gotNop bool
	handler := func(ec echo.Context) error {
		l := fctx.Logger(fctx.From(ec))
		l.Info("test") // must not panic
		gotNop = l != f.logger
		return nil
	}

	err := fctx.Middleware(f.logger, f.tracer)(destroyingMW(handler))(f.ec)
	require.NoError(t, err)
	assert.True(t, gotNop, "values lost, Logger returns nop fallback")
}
