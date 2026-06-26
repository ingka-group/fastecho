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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestMetricsEndpointStillServed(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")

	fe, err := Initialize(&Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = fe.Shutdown(context.Background()) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	fe.Handler().ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)

	// A scraper negotiating OpenMetrics must get OpenMetrics back: that is the
	// exposition format that carries exemplars (trace_id/span_id). Without
	// HandlerOpts.EnableOpenMetrics, promhttp ignores the Accept header and returns
	// plain text, dropping exemplars - so this guards that the flag stays set.
	om := httptest.NewRecorder()
	omReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	omReq.Header.Set("Accept", "application/openmetrics-text")
	fe.Handler().ServeHTTP(om, omReq)
	assert.Equal(t, 200, om.Code)
	assert.Contains(t, om.Header().Get("Content-Type"), "application/openmetrics-text")
}

func TestMetricsEndpointNotMountedWithoutPrometheus(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none") // not prometheus => nil gatherer => no /metrics

	fe, err := Initialize(&Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = fe.Shutdown(context.Background()) })

	rec := httptest.NewRecorder()
	fe.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code, "/metrics is not mounted when the exporter isn't prometheus")
}

// TestOtelecho_RecordsRoutePattern exercises otelecho directly with a recorder.
func TestOtelecho_RecordsRoutePattern(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	e := echo.New()
	e.Use(otelecho.Middleware("test", otelecho.WithTracerProvider(tp)))
	e.GET("/widgets/:id", func(c echo.Context) error { return c.String(200, "ok") })

	r := httptest.NewRecorder()
	e.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/widgets/42", nil))

	spans := rec.Ended()
	require.Len(t, spans, 1)
	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes() {
		attrs[string(kv.Key)] = kv.Value.String()
	}
	assert.Equal(t, "/widgets/:id", attrs["http.route"])
	assert.Equal(t, "GET", attrs["http.request.method"])
}
