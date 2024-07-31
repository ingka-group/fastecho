package core

import (
	gocontext "context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/config"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/context"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/env"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/router"
	"github.com/ingka-group-digital/ocp-go-utils/echozap"
)

// ServerConfig serves as input configuration for the service
type ServerConfig struct {
	ExtraEnvVars        env.EnvVars
	ValidationRegistrar func(v *router.Validator) error
	Routes              []router.Route
	ContextProps        any
	Options             config.Options
}

// server is a wrapper around Echo
type server struct {
	Echo      *echo.Echo
	Validator *router.Validator

	// Not accessible from another package
	router         *router.Router
	tracerProvider *sdktrace.TracerProvider
}

func newServer() *server {
	return &server{}
}

// NewServer returns a new instance of Server which contains an Echo server
func NewServer(cfg ServerConfig) error {
	// set up the server
	s := newServer()
	envs, err := s.setup(cfg)
	if err != nil {
		return err
	}

	// run it!
	return s.run(envs[env.Hostname].Value, envs[env.Port].Value)
}

// setup sets up the service with the given environment variables and an optional postgres db layer
func (s *server) setup(serverCfg ServerConfig) (env.EnvVars, error) {
	var err error

	// set up echo
	s.Echo = echo.New()

	// set up validation
	s.Validator, err = router.NewValidator()
	if err != nil {
		return nil, err
	}

	// register custom validations
	err = serverCfg.ValidationRegistrar(s.Validator)
	if err != nil {
		return nil, err
	}
	s.Echo.Validator = s.Validator

	// set up service config
	var cfg *config.ServiceConfig
	cfg, s.tracerProvider, err = config.NewServiceConfig(serverCfg.ExtraEnvVars, serverCfg.Options)
	if err != nil {
		return nil, err
	}

	// set up middlewares
	configureMiddlewares(s.Echo, cfg.Logger, serverCfg.Options.SkipMetrics, cfg.Tracer, serverCfg.ContextProps)

	// set up routes
	s.router = router.NewRouter(serverCfg.Routes, serverCfg.Options)
	err = s.router.RegisterRoutes(s.Echo, cfg.Env)
	if err != nil {
		return nil, err
	}

	return cfg.Env, err
}

// configureMiddlewares configures all the middlewares for Echo.
func configureMiddlewares(e *echo.Echo, logger *zap.Logger, SkipMetrics bool, tracer *trace.Tracer, props any) {
	// CORS support
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Request ID
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx) || isHealthRoute(ctx)
		},
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	// Zap Logger
	e.Use(echozap.ZapLoggerMiddlewareWithConfig(logger, echozap.ZapLoggerMiddlewareConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx) || isHealthRoute(ctx)
		},
	}))

	// Context
	e.Use(context.ServiceContextMiddleware(logger, tracer, props))

	// Gzip
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(ctx echo.Context) bool {
			return isSwaggerRoute(ctx) || isMetricsRoute(ctx)
		},
	}))

	if !SkipMetrics {
		// Metrics
		e.Use(echoprometheus.NewMiddleware("echo_http"))
	}

	// Recover
	e.Use(middleware.Recover())
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

// Run starts the server and listens for interrupt signals to gracefully shut down the server
func (s *server) run(host string, port string) error {
	// defer the shutdown of the tracer provider
	defer func() {
		if s.tracerProvider != nil {
			_ = s.tracerProvider.Shutdown(gocontext.Background())
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
