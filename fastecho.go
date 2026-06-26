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
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/echozap"
	"github.com/ingka-group/fastecho/env"
	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/router"
	"github.com/ingka-group/fastecho/telemetry"
)

const (
	hostname        = "HOSTNAME"
	port            = "PORT"
	swaggerUITitle  = "SWAGGER_UI_TITLE"
	swaggerJSONPath = "SWAGGER_JSON_PATH"
)

const (
	defaultWorkerInitialRestartDelay  = 1 * time.Second
	defaultWorkerMaxRestartDelay      = 30 * time.Second
	defaultWorkerStableResetThreshold = 1 * time.Minute
	defaultWorkerCrashLoopThreshold   = 10
)

var (
	// Environment variables for fastecho to operate.
	envs = env.Map{
		hostname: {
			DefaultValue: "localhost",
		},
		port: {
			DefaultValue: "8080",
			IsInteger:    true,
		},
		swaggerJSONPath: {
			// Defines the path to the swagger.json file on the server. This is used by the swagger UI.
			DefaultValue: "/swagger/swagger.json",
		},
		swaggerUITitle: {
			DefaultValue: "FastEcho Service",
		},
		env.LogLevel: {
			DefaultValue: env.DevLogLevel,
			OneOf:        []string{env.DevLogLevel, env.TestLogLevel, env.ProdLogLevel},
		},
	}
)

// server is a wrapper around Echo.
type server struct {
	Echo      *echo.Echo
	Router    *router.Router
	Logger    *zap.Logger
	Providers *telemetry.Providers
	Workers   map[string]Worker

	workerInitialRestartDelay  time.Duration
	workerMaxRestartDelay      time.Duration
	workerStableResetThreshold time.Duration
	workerCrashLoopThreshold   int
	workerFailures             metric.Int64Counter
}

type FastEcho struct {
	server *server
}

// Run starts a new instance of fastecho.
func Run(cfg *Config) error {
	s, err := newServer(cfg)
	if err != nil {
		return err
	}

	// Allow custom Echo configuration
	if cfg.EchoFn != nil {
		err = cfg.EchoFn(s.Echo)
		if err != nil {
			return err
		}
	}

	err = s.Router.Setup()
	if err != nil {
		return err
	}

	s.Router.PrintRoutes(s.Echo)

	// Run it!
	return s.run(envs[hostname].Value, envs[port].Value)
}

// Initialize sets up a new instance of FastEcho and returns a prepared FastEcho type, but does not
// boot the server.
func Initialize(cfg *Config) (*FastEcho, error) {
	s, err := newServer(cfg)
	if err != nil {
		return nil, err
	}

	return &FastEcho{server: s}, nil
}

// Handler returns the Echo handler for the defined FastEcho server.
func (fe *FastEcho) Handler() http.Handler {
	return fe.server.Echo
}

// Shutdown cleanly shuts down the server and any telemetry providers. Both are
// always attempted and their errors joined - a telemetry-flush error must not
// skip draining in-flight HTTP requests.
func (fe *FastEcho) Shutdown(ctx context.Context) error {
	provErr := fe.server.Providers.Shutdown(ctx)
	echoErr := fe.server.Echo.Shutdown(ctx)
	return errors.Join(provErr, echoErr)
}

// BindValidate binds the request body to v and validates it using the
// registered validator. Use this at the handler boundary.
func BindValidate(ec echo.Context, v any) error {
	if err := ec.Bind(v); err != nil {
		return err
	}
	return ec.Validate(v)
}

func newServer(cfg *Config) (*server, error) {
	// Set up the server
	s := &server{
		workerInitialRestartDelay:  defaultWorkerInitialRestartDelay,
		workerMaxRestartDelay:      defaultWorkerMaxRestartDelay,
		workerStableResetThreshold: defaultWorkerStableResetThreshold,
		workerCrashLoopThreshold:   defaultWorkerCrashLoopThreshold,
	}

	// If no configuration is passed,
	// the service should still run with default values
	if cfg == nil {
		cfg = &Config{}
	}

	err := s.setup(cfg)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// setup sets up the service with the given environment variables and an optional postgres db layer
func (s *server) setup(cfg *Config) error {
	var err error

	// set up echo
	s.Echo = echo.New()

	// config the service
	err = s.config(cfg)
	if err != nil {
		return err
	}

	// set up middlewares
	s.middlewares(cfg)

	// Print log level configuration at startup
	logLevel := env.GetLogLevel()
	printBanner("fastecho log configuration",
		"LOG_LEVEL (env)", logLevel,
		"EchoZap level", s.Logger.Level().String(),
	)

	fastechoRouter, err := router.NewRouter(
		router.Config{
			Echo:             s.Echo,
			Routes:           cfg.Routes,
			SkipMetrics:      cfg.Opts.Metrics.Skip,
			SkipHealthChecks: cfg.Opts.HealthChecks.Skip,
			HealthChecksDB:   cfg.Opts.HealthChecks.DB,
			SwaggerTitle:     envs[swaggerUITitle].Value,
			SwaggerPath:      envs[swaggerJSONPath].Value,
			MetricsGatherer:  s.Providers.PrometheusGatherer,
		},
	)
	if err != nil {
		return err
	}

	// set up validation
	vdt, err := router.NewValidator()
	if err != nil {
		return err
	}
	if cfg.ValidationRegistrar != nil {
		// register custom validations
		err = cfg.ValidationRegistrar(vdt)
		if err != nil {
			return err
		}
	}
	// register plugin validations and routes
	for _, plugin := range cfg.Plugins {
		if plugin.ValidationRegistrar != nil {
			err = plugin.ValidationRegistrar(vdt)
			if err != nil {
				return errors.New("error registering plugin validation: " + err.Error())
			}
		}
		// Register plugin routes
		fmt.Println("Registering plugin routes")
		err = plugin.Routes(s.Echo, fastechoRouter)
		if err != nil {
			return errors.New("error registering plugin routes: " + err.Error())
		}
	}
	s.Echo.Validator = vdt
	s.Router = fastechoRouter
	s.Workers = cfg.Workers

	return err
}

func (s *server) config(cfg *Config) error {
	// Set environment variables MUST be the first step
	// merge default env vars with extra env vars

	var allEnvs = make(env.Map)
	maps.Copy(allEnvs, envs)
	maps.Copy(allEnvs, cfg.ExtraEnvs)

	err := allEnvs.SetEnv()
	if err != nil {
		return err
	}

	logger, err := echozap.New()
	if err != nil {
		return err
	}
	s.Logger = logger

	providers, info, err := telemetry.Init(context.Background(), telemetry.Config{
		SetGlobal:   true,
		SkipTraces:  cfg.Opts.Tracing.Skip,
		SkipMetrics: cfg.Opts.Metrics.Skip,
	})
	if err != nil {
		return err
	}
	s.Providers = providers

	endpoint := info.OTLPEndpoint
	if endpoint == "" {
		endpoint = "(SDK default)"
	}
	s.Logger.Info("telemetry configured",
		zap.String("service_name", info.ServiceName),
		zap.Bool("traces", info.Traces),
		zap.String("traces_exporter", info.TracesExporter),
		zap.String("otlp_protocol", info.OTLPProtocol),
		zap.String("otlp_endpoint", endpoint),
		zap.Bool("metrics", info.Metrics),
		zap.String("metrics_exporter", info.MetricsExporter),
		zap.String("metrics_delivery", info.MetricsDelivery),
	)

	return nil
}

// middlewares configures all the middlewares for Echo.
func (s *server) middlewares(cfg *Config) {
	skip := func(c echo.Context) bool {
		return isSwaggerRoute(c) || isMetricsRoute(c) || isHealthRoute(c)
	}

	// 1. Recover, outermost: catches panics in any middleware below.
	s.Echo.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			req := c.Request()
			fields := append(fctx.Fields(req.Context()),
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
				zap.String("uri", req.RequestURI),
				zap.ByteString("stack", stack),
			)
			s.Logger.Error("panic recovered", fields...)
			return err
		},
	}))

	// 2. Request ID, before trace: so it can be set as a span attribute.
	s.Echo.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper:   skip,
		Generator: fctx.NewRequestID,
	}))

	// 3. otelecho trace + metrics, registered only when a signal is on.
	if !cfg.Opts.Tracing.Skip || !cfg.Opts.Metrics.Skip {
		s.Echo.Use(otelecho.Middleware(
			otelServiceName(),
			otelecho.WithTracerProvider(s.Providers.TracerProvider),
			otelecho.WithMeterProvider(s.Providers.MeterProvider),
			otelecho.WithSkipper(skip),
		))
	}

	// 4. fctx: enriches the request logger + sets fastecho.request_id on the span.
	// Providers is always non-nil; a noop tracer when tracing is skipped.
	tracer := s.Providers.TracerProvider.Tracer(telemetry.ScopeName)
	s.Echo.Use(fctx.Middleware(s.Logger, tracer))

	// 5. Gzip.
	s.Echo.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool { return isSwaggerRoute(c) || isMetricsRoute(c) },
	}))

	// 6. Access log, innermost: sees the final status via the request logger.
	if !cfg.Opts.Logs.Skip {
		s.Echo.Use(echozap.ZapLoggerMiddlewareWithConfig(s.Logger, echozap.ZapLoggerMiddlewareConfig{
			Skipper: skip,
		}))
	}
}

// run starts the server and listens for interrupt signals to gracefully shut it down.
func (s *server) run(host string, port string) error {
	// Catch-all for crashes the runtime can still report: on a runtime-fatal
	// error (stack overflow, concurrent map writes, out-of-memory, etc.) or an
	// unrecovered panic, the runtime writes a full stack dump to stderr before
	// exiting. "all" includes every goroutine, not just the offending one, so a
	// restarted service always leaves a traceable cause in the logs. Mirrors
	// GOTRACEBACK=all without relying on the deployment to set it. Note: this
	// cannot help on SIGKILL / kernel OOM-kill, where the process dies instantly.
	debug.SetTraceback("all")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return s.serve(ctx, host, port)
}

// serve starts the HTTP server and shuts it down gracefully once ctx is done.
func (s *server) serve(ctx context.Context, host string, port string) error {
	// Defer the shutdown of the telemetry providers
	defer func() { _ = s.Providers.Shutdown(context.Background()) }()

	// Flush buffered logs on the way out
	defer func() { _ = s.Logger.Sync() }()

	// One counter for all workers; labelled per worker/kind at record time.
	// MeterProvider is always non-nil (noop when metrics are skipped → Add is a
	// no-op), so this is safe to create unconditionally.
	s.workerFailures, _ = s.Providers.MeterProvider.Meter(telemetry.ScopeName).Int64Counter(
		"fastecho.worker.failures",
		metric.WithDescription("background worker panics and error exits"),
		metric.WithUnit("{failure}"),
	)

	// Start background workers, tracked so shutdown can wait for them to drain.
	var workers sync.WaitGroup
	for name, w := range s.Workers {
		workers.Go(func() { s.runWorker(ctx, name, w) })
	}

	// Start server
	go func() {
		serviceURL := fmt.Sprintf("%s:%v", host, port)
		if err := s.Echo.Start(serviceURL); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Echo.Logger.Panicf("Shutting down the server! \n%s", err)
		}
	}()

	// Wait for the shutdown signal, then drain in-flight requests with a 10s timeout.
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Echo.Shutdown(shutdownCtx)

	// Wait for workers to unwind, bounded by the same shutdown budget.
	s.drainWorkers(shutdownCtx, &workers)

	return err
}

// isMetricsRoute returns whether the request is to metrics endpoint.
func isMetricsRoute(ctx echo.Context) bool {
	return strings.Contains(ctx.Request().URL.Path, "/metrics")
}

// isSwaggerRoute returns whether the request is to swagger endpoint.
func isSwaggerRoute(ctx echo.Context) bool {
	return strings.Contains(ctx.Request().URL.Path, "/swagger/")
}

// isHealthRoute returns whether the request is to health endpoint.
func isHealthRoute(ctx echo.Context) bool {
	return strings.Contains(ctx.Request().URL.Path, "/health")
}

// otelServiceName is the span/metric service label otelecho requires as its
// first arg. The resource service.name comes from OTEL_SERVICE_NAME; reuse it.
func otelServiceName() string {
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		return v
	}
	return "fastecho"
}
