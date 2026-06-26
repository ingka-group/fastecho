// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
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

package fastecho

import (
	gocontext "context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/router"
	"github.com/ingka-group/fastecho/telemetry"
)

// newTestServer builds a bare server with a no-op logger, enough to test
// startup and shutdown.
func newTestServer() *server {
	return &server{
		Echo:   echo.New(),
		Logger: zap.NewNop(),
		Providers: &telemetry.Providers{
			TracerProvider: tracenoop.NewTracerProvider(),
			MeterProvider:  metricnoop.NewMeterProvider(),
		},

		workerInitialRestartDelay:  1 * time.Millisecond,
		workerMaxRestartDelay:      10 * time.Millisecond,
		workerStableResetThreshold: 1 * time.Second,
		workerCrashLoopThreshold:   10,
	}
}

// freePort grabs a free port and returns it. Tests use it so they can wait
// for the server to come up instead of guessing with a sleep.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	p := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	require.NoError(t, ln.Close())
	return p
}

// waitUntilServing waits until the server accepts a connection on addr. In
// run() the signal handler is set up before the server starts listening, so
// once we can connect we know the handler is ready too.
func waitUntilServing(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("server did not start listening on %s within timeout", addr)
}

// waitForReturn fails the test if the server does not stop in time.
func waitForReturn(t *testing.T, errc <-chan error) error {
	t.Helper()
	select {
	case err := <-errc:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
		return nil
	}
}

func TestServeShutsDownWhenContextCancelled(t *testing.T) {
	s := newTestServer()
	p := freePort(t)
	ctx, cancel := gocontext.WithCancel(gocontext.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	waitUntilServing(t, net.JoinHostPort("localhost", p))
	cancel()

	assert.NoError(t, waitForReturn(t, errc))
}

func TestRunShutsDownOnSignal(t *testing.T) {
	// Signals hit the whole process, so run these one at a time (no
	// t.Parallel): each run() adds a handler and removes it when it returns.
	signals := map[string]syscall.Signal{
		"sigterm": syscall.SIGTERM,
		"sigint":  syscall.SIGINT,
	}

	for name, sig := range signals {
		t.Run(name, func(t *testing.T) {
			s := newTestServer()
			p := freePort(t)

			errc := make(chan error, 1)
			go func() {
				errc <- s.run("localhost", p)
			}()

			waitUntilServing(t, net.JoinHostPort("localhost", p))

			pr, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			require.NoError(t, pr.Signal(sig))

			assert.NoError(t, waitForReturn(t, errc))
		})
	}
}

func TestServeWaitsForInFlightRequest(t *testing.T) {
	s := newTestServer()

	started := make(chan struct{})
	s.Echo.GET("/slow", func(c echo.Context) error {
		close(started)
		time.Sleep(200 * time.Millisecond)
		return c.String(http.StatusOK, "done")
	})

	p := freePort(t)
	ctx, cancel := gocontext.WithCancel(gocontext.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	addr := net.JoinHostPort("localhost", p)
	waitUntilServing(t, addr)

	respc := make(chan *http.Response, 1)
	reqErrc := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			reqErrc <- err
			return
		}
		respc <- resp
	}()

	// Start shutdown only after the request has reached the handler.
	<-started
	cancel()

	select {
	case resp := <-respc:
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "done", string(body))
	case err := <-reqErrc:
		t.Fatalf("in-flight request was dropped instead of drained: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("in-flight request did not complete")
	}

	assert.NoError(t, waitForReturn(t, errc))
}

func TestServeFlushesLogsOnShutdown(t *testing.T) {
	rec := &recordingSyncer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(rec),
		zapcore.InfoLevel,
	)
	s := &server{
		Echo:   echo.New(),
		Logger: zap.New(core),
		Providers: &telemetry.Providers{
			TracerProvider: tracenoop.NewTracerProvider(),
			MeterProvider:  metricnoop.NewMeterProvider(),
		},
	}

	p := freePort(t)
	ctx, cancel := gocontext.WithCancel(gocontext.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	waitUntilServing(t, net.JoinHostPort("localhost", p))
	cancel()

	require.NoError(t, waitForReturn(t, errc))
	assert.True(t, rec.didSync(), "expected Logger.Sync to be called on the shutdown path")
}

// recordingSyncer is a fake log writer that just remembers if Sync was called.
type recordingSyncer struct {
	mu     sync.Mutex
	synced bool
}

func (r *recordingSyncer) Write(p []byte) (int, error) { return len(p), nil }

func (r *recordingSyncer) Sync() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.synced = true
	return nil
}

func (r *recordingSyncer) didSync() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.synced
}

func TestInitialize_HasProvidersAndShutdownIsSafe(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	fe, err := Initialize(&Config{})
	require.NoError(t, err)
	require.NotNil(t, fe.server.Providers)
	require.NoError(t, fe.Shutdown(gocontext.Background()))
}

func TestMetricsEndpointStillServed(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")

	fe, err := Initialize(&Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = fe.Shutdown(gocontext.Background()) })

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
	t.Cleanup(func() { _ = fe.Shutdown(gocontext.Background()) })

	rec := httptest.NewRecorder()
	fe.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code, "/metrics is not mounted when the exporter isn't prometheus")
}

// TestOtelecho_RecordsRoutePattern exercises otelecho directly with a recorder.
func TestOtelecho_RecordsRoutePattern(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	t.Cleanup(func() { _ = tp.Shutdown(gocontext.Background()) })

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
	t.Cleanup(func() { _ = fe.Shutdown(gocontext.Background()) })

	rec := httptest.NewRecorder()
	fe.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	// recover turns the panic into a 500 instead of crashing the server.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServerSpan_CarriesRequestID(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	t.Cleanup(func() { _ = tp.Shutdown(gocontext.Background()) })

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
