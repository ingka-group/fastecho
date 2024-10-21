package fastecho

import (
	gocontext "context"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/echozap"
	"github.com/ingka-group-digital/ocp-go-utils/fastecho/context"
	"github.com/ingka-group-digital/ocp-go-utils/fastecho/env"
	"github.com/ingka-group-digital/ocp-go-utils/fastecho/errs"
	"github.com/ingka-group-digital/ocp-go-utils/fastecho/router"
	"github.com/ingka-group-digital/ocp-go-utils/otel"
	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

const (
	hostname = "HOSTNAME"
	port     = "PORT"
	envType  = "ENV_TYPE"

	swaggerUITitle  = "SWAGGER_UI_TITLE"
	swaggerJSONPath = "SWAGGER_JSON_PATH"

	localEnv = "local"
	devEnv   = "dev"
	testEnv  = "test"
	prodEnv  = "prod"
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
		envType: {
			DefaultValue: devEnv,
			OneOf:        []string{localEnv, devEnv, testEnv, prodEnv},
		},
		swaggerJSONPath: {
			// Defines the path to the swagger.json file on the server. This is used by the swagger UI.
			DefaultValue: "/swagger/swagger.json",
		},
		swaggerUITitle: {
			DefaultValue: "FastEcho Service",
		},
	}
)

// Config serves as input configuration for fastecho.
type Config struct {
	ExtraEnvs           env.Map
	ValidationRegistrar func(v *router.Validator) error
	Routes              func(e *echo.Echo, r *router.Router) error
	ContextProps        any
	Opts                Opts
}

// Opts define configuration options for fastecho.
type Opts struct {
	Metrics      MetricsOpts
	Tracing      TracingOpts
	HealthChecks HealthChecksOpts
}

// MetricsOpts define configuration options for metrics.
type MetricsOpts struct {
	Skip bool
}

// TracingOpts define configuration options for tracing.
type TracingOpts struct {
	Skip        bool
	ServiceName string
}

// HealthChecksOpts define configuration options for health checks.
type HealthChecksOpts struct {
	Skip bool
	DB   *gorm.DB
}

// server is a wrapper around Echo.
type server struct {
	Echo           *echo.Echo
	Router         *router.Router
	Logger         *zap.Logger
	Tracer         *trace.Tracer
	TracerProvider *sdktrace.TracerProvider
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

// Shutdown cleanly shuts down the server and any tracing providers.
func (fe *FastEcho) Shutdown(ctx gocontext.Context) error {
	if fe.server.TracerProvider != nil {
		_ = fe.server.TracerProvider.Shutdown(ctx)
	}
	// Clean up global variables after shutdown
	defer func() { prometheus.DefaultRegisterer = prometheus.NewRegistry() }()

	return fe.server.Echo.Shutdown(ctx)
}

func newServer(cfg *Config) (*server, error) {
	// Set up the server
	s := &server{}

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
	s.Echo.Validator = vdt

	// config the service
	err = s.config(cfg)
	if err != nil {
		return err
	}

	// set up middlewares
	s.middlewares(cfg)

	router, err := router.NewRouter(
		router.Config{
			Echo:             s.Echo,
			Routes:           cfg.Routes,
			SkipMetrics:      cfg.Opts.Metrics.Skip,
			SkipHealthChecks: cfg.Opts.HealthChecks.Skip,
			HealthChecksDB:   cfg.Opts.HealthChecks.DB,
			SwaggerTitle:     envs[swaggerUITitle].Value,
			SwaggerPath:      envs[swaggerJSONPath].Value,
		},
	)
	if err != nil {
		return err
	}
	s.Router = router

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

	// only init tracing if it's not disabled
	var tracerProvider *sdktrace.TracerProvider
	var tracer *trace.Tracer

	// enable/disable tracing
	if !cfg.Opts.Tracing.Skip {
		if stringutils.IsEmpty(cfg.Opts.Tracing.ServiceName) {
			return errs.New("service name not provided for tracing")
		}

		tracerProvider, tracer, err = newTracer(cfg.Opts.Tracing.ServiceName)
		if err != nil {
			return err
		}

		s.Tracer = tracer
		s.TracerProvider = tracerProvider
	}

	return nil
}

// middlewares configures all the middlewares for Echo.
func (s *server) middlewares(cfg *Config) {
	// CORS support
	s.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	if !cfg.Opts.Tracing.Skip {
		s.Echo.Use(otel.Middleware(
			otel.WithSkipper(func(ctx echo.Context) bool {
				return isSwaggerRoute(ctx) || isMetricsRoute(ctx) || isHealthRoute(ctx)
			}),
			otel.WithServiceName(cfg.Opts.Tracing.ServiceName),
			otel.WithEnv(envs[envType].Value),
		))
	}

	// Request ID
	s.Echo.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx) || isHealthRoute(ctx)
		},
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	// Zap Logger
	s.Echo.Use(echozap.ZapLoggerMiddlewareWithConfig(s.Logger, echozap.ZapLoggerMiddlewareConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx) || isHealthRoute(ctx)
		},
	}))

	// Context
	s.Echo.Use(context.ServiceContextMiddleware(s.Logger, s.Tracer, cfg.ContextProps))

	// Gzip
	s.Echo.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx)
		},
	}))

	// Metrics
	if !cfg.Opts.Metrics.Skip {
		s.Echo.Use(echoprometheus.NewMiddleware("echo_http"))
	}

	// Recover
	s.Echo.Use(middleware.Recover())
}

// run starts the server and listens for interrupt signals to gracefully shut it down.
func (s *server) run(host string, port string) error {
	// Defer the shutdown of the tracer provider
	defer func() {
		if s.TracerProvider != nil {
			_ = s.TracerProvider.Shutdown(gocontext.Background())
		}
	}()

	// Start server
	go func() {
		serviceURL := fmt.Sprintf("%s:%v", host, port)
		if err := s.Echo.Start(serviceURL); err != nil && err != http.ErrServerClosed {
			s.Echo.Logger.Panicf("Shutting down the server! \n%s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 10*time.Second)
	defer cancel()

	err := s.Echo.Shutdown(ctx)
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
