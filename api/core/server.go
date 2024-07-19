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
	"gorm.io/gorm"

	"github.com/ingka-group-digital/ocp-go-utils/api/core/config"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/context"
	"github.com/ingka-group-digital/ocp-go-utils/api/core/router"
	"github.com/ingka-group-digital/ocp-go-utils/echozap"
)

// Server is a wrapper around Echo
type Server struct {
	Echo      *echo.Echo
	Database  *gorm.DB
	Validator *router.Validator

	// Not accessible from another package
	tracerProvider *sdktrace.TracerProvider
}

// NewServer returns a new instance of Server which contains an Echo server
func NewServer(extraEnvVars *config.EnvVar, props *map[string]interface{}, withPostgres bool) (*Server, error) {
	s := &Server{}
	err := s.setup(extraEnvVars, props, withPostgres)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// setup sets up the service with the given environment variables and an optional postgres db layer
func (s *Server) setup(extraEnvVars *config.EnvVar, props *map[string]interface{}, withPostgres bool) error {
	var err error

	s.Echo = echo.New()

	s.Validator, err = router.NewValidator()
	if err != nil {
		return err
	}

	var cfg *config.ServiceConfig
	cfg, s.tracerProvider, err = config.NewServiceConfig(extraEnvVars, withPostgres)
	if err != nil {
		return err
	}

	// Set the database to be picked up by the caller
	s.Database = cfg.Database

	configureMiddlewares(s.Echo, cfg.Logger, cfg.Tracer, props)

	return nil
}

// configureMiddlewares configures all the middlewares for Echo.
func configureMiddlewares(e *echo.Echo, logger *zap.Logger, tracer *trace.Tracer, props *map[string]interface{}) {
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

	// Metrics
	e.Use(echoprometheus.NewMiddleware("echo_http"))

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
func (s *Server) Run() error {
	// defer the shutdown of the tracer provider
	defer func() {
		if s.tracerProvider != nil {
			_ = s.tracerProvider.Shutdown(gocontext.Background())
		}
	}()

	// Start server
	go func() {
		serviceURL := fmt.Sprintf("%s:%v", config.Env[config.Hostname].Value, config.Env[config.Port].Value)
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
