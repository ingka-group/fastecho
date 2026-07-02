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
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ingka-group/fastecho/fctx"
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

			var lvl zapcore.Level
			var msg string
			switch n := res.Status; {
			case n >= 500:
				lvl, msg = zapcore.ErrorLevel, "Server error"
			case n >= 400:
				lvl, msg = zapcore.WarnLevel, "Client error"
			case n >= 300:
				lvl, msg = zapcore.InfoLevel, "Redirection"
			default:
				lvl, msg = zapcore.DebugLevel, "Success"
			}
			// Check before building any field: at prod level every 2xx line is
			// dropped, and this keeps that hot path allocation-free.
			ce := log.Check(lvl, msg)
			if ce == nil {
				return nil
			}

			fields := append(fctx.Fields(req.Context()),
				zap.String("remote_ip", c.RealIP()),
				zap.Float64("latency_ms", float64(time.Since(start))/float64(time.Millisecond)),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.String("path", c.Path()),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			)
			if fctx.RequestID(req.Context()) == "" {
				if reqID := res.Header().Get(echo.HeaderXRequestID); reqID != "" {
					fields = append(fields, zap.String("request_id", reqID))
				}
			}
			if res.Status >= 400 {
				// zap.Error(nil) is a no-op field, so a handler writing the
				// status directly doesn't add an empty error.
				fields = append(fields, zap.Error(err))
			}
			ce.Write(fields...)

			return nil
		}
	}
}
