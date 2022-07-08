package echozap

import (
	"fmt"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	EnvTypeKey = "ENV_TYPE"
	DevEnv     = "dev"
	TestEnv    = "test"
	ProdEnv    = "prod"
)

type (
	Skipper func(c echo.Context) bool

	// ZapLoggerMiddlewareConfig defines the config for ZapLogger middleware
	ZapLoggerMiddlewareConfig struct {
		// Skipper defines a function to skip middleware
		Skipper Skipper
	}
)

var (
	// DefaultZapLoggerMiddlewareConfig is the default ZapLogger middleware config
	DefaultZapLoggerMiddlewareConfig = ZapLoggerMiddlewareConfig{
		Skipper: DefaultSkipper,
	}
)

// DefaultSkipper returns false which processes the middleware
func DefaultSkipper(echo.Context) bool {
	return false
}

// ZapLoggerMiddleware is a middleware for zap to provide an "access log" like logging for each request
func ZapLoggerMiddleware(log *zap.Logger) echo.MiddlewareFunc {
	return ZapLoggerMiddlewareWithConfig(log, DefaultZapLoggerMiddlewareConfig)
}

// ZapLoggerMiddlewareWithConfig is a middleware (with configuration) for zap to provide an "access log" like logging for each request
//
// This is an extended version from library https://github.com/brpaz/echozap to use a Skipper
func ZapLoggerMiddlewareWithConfig(log *zap.Logger, config ZapLoggerMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Defaults
		if config.Skipper == nil {
			config.Skipper = DefaultZapLoggerMiddlewareConfig.Skipper
		}

		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
				fields = append(fields, zap.String("request_id", id))
			}

			n := res.Status
			switch {
			case n >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}

// getEnvType() get the env type from the OS env
func getEnvType() (string, error) {
	envType := os.Getenv(EnvTypeKey)

	if envType != DevEnv && envType != TestEnv && envType != ProdEnv {
		err := fmt.Errorf("please set %s to %s, %s or %s", EnvTypeKey, DevEnv, TestEnv, ProdEnv)
		return "", err
	}

	return envType, nil
}

// New provides a logger with sain defaults for logging to server ENVs (dev,test,prod)
// It configures a json structured logger that writes info messages to stdout
func New() (*zap.Logger, error) {
	envType, err := getEnvType()
	if err != nil {
		return nil, err
	}

	var config zap.Config
	if envType == ProdEnv {
		config = zap.NewProductionConfig()
	} else { // TestEnv, DevEnv
		config = zap.NewDevelopmentConfig()

		// Custom zap.NewDevelopmentConfig settings
		config.EncoderConfig = zap.NewProductionEncoderConfig()
		config.Encoding = "json" // Use structure logging
	}

	//Use CapitalLevelEncoder in all envs
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	//Make sure info level messages are written to stdout in all envs
	config.OutputPaths = []string{"stdout"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	defer func(zapLogger *zap.Logger) {
		_ = zapLogger.Sync()
	}(zapLogger)

	return zapLogger, nil
}
