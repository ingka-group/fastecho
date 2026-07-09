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

package telemetry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/telemetry"
)

func TestWrapClient_InjectsTraceparentAndRequestID(t *testing.T) {
	// otelhttp injects via the global propagator, which defaults to a no-op;
	// without this the traceparent assertion below fails (nothing is injected).
	// Restore the prior propagator so this global mutation doesn't leak to other tests.
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() { otel.SetTextMapPropagator(prevProp) })
	otel.SetTextMapPropagator(propagation.TraceContext{})

	var gotTraceparent, gotRequestID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
		gotRequestID = r.Header.Get("X-Request-Id")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	// Caller brings a normal client; WrapClient only augments its transport.
	client := telemetry.WrapClient(&http.Client{})

	ctx := fctx.WithRequestID(t.Context(), "req-xyz")
	// Start a span so there is a traceparent to inject.
	ctx = seedSpan(t, ctx)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	assert.NotEmpty(t, gotTraceparent, "traceparent injected")
	assert.Equal(t, "req-xyz", gotRequestID, "X-Request-Id forwarded from fctx")
}

func seedSpan(t *testing.T, ctx context.Context) context.Context {
	t.Helper()
	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	ctx, span := tp.Tracer("test").Start(ctx, "client-call")
	t.Cleanup(func() { span.End() })
	return ctx
}

func TestRoundTrip_PropagatesTraceAndRequestID(t *testing.T) {
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() { otel.SetTextMapPropagator(prevProp) })
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	// Downstream fastecho-like server: otelecho continues the inbound trace and
	// echoes the request id + the trace id it saw.
	e := echo.New()
	e.Use(otelecho.Middleware("downstream", otelecho.WithTracerProvider(tp)))
	e.GET("/echo", func(c echo.Context) error {
		sc := trace.SpanContextFromContext(c.Request().Context())
		return c.JSON(200, map[string]string{
			"request_id": c.Request().Header.Get("X-Request-Id"),
			"trace_id":   sc.TraceID().String(),
		})
	})
	srv := httptest.NewServer(e)
	defer srv.Close()

	ctx, span := tp.Tracer("caller").Start(t.Context(), "outbound")
	defer span.End()
	ctx = fctx.WithRequestID(ctx, "rid-roundtrip")

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/echo", nil)
	resp, err := telemetry.WrapClient(&http.Client{}).Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var got map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "rid-roundtrip", got["request_id"], "X-Request-Id forwarded")
	assert.Equal(t, span.SpanContext().TraceID().String(), got["trace_id"], "trace continued across the hop")
}

// Wrapping twice must not nest another otelhttp layer (duplicate client spans
// and duplicated headers); WrapTransport is idempotent.
func TestWrapTransport_Idempotent(t *testing.T) {
	w := telemetry.WrapTransport(nil)
	assert.Equal(t, w, telemetry.WrapTransport(w), "wrapping an already-wrapped transport is a no-op")
}

// closableTransport records whether CloseIdleConnections reached it.
type closableTransport struct {
	http.RoundTripper
	closed bool
}

func (c *closableTransport) CloseIdleConnections() { c.closed = true }

// net/http type-asserts Client.Transport for CloseIdleConnections; the wrapper
// must forward it to the original base (otelhttp forwards nothing).
func TestWrapClient_ForwardsCloseIdleConnections(t *testing.T) {
	base := &closableTransport{RoundTripper: http.DefaultTransport}
	client := telemetry.WrapClient(&http.Client{Transport: base})

	client.CloseIdleConnections()

	assert.True(t, base.closed, "CloseIdleConnections must reach the wrapped base transport")
}
