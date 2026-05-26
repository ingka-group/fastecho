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

package context

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/fctx"
)

// Deprecated: Use [fctx.From] and [fctx.Logger]/[fctx.Tracer]/[fctx.RequestID].
// For shared service configuration, use constructor injection.
type ServiceContext[T any] struct {
	echo.Context
	ZapLogger    *zap.Logger
	Tracer       *trace.Tracer
	Props        T
	RequestProps map[string]interface{}
}

// Deprecated: Use [fastecho.BindValidate] instead.
func (c *ServiceContext[T]) BindValidate(i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}

	if err := c.Validate(i); err != nil {
		return err
	}

	return nil
}

// Deprecated: Use [fctx.From] plus [fctx.Logger]/[fctx.Tracer].
// Under [fctx.Middleware], returns a best-effort shim from fctx keys.
func GetServiceContext[T any](ctx echo.Context) *ServiceContext[T] {
	if sctx, ok := ctx.(*ServiceContext[T]); ok {
		return sctx
	}

	// Compatibility shim for new middleware
	reqCtx := ctx.Request().Context()
	var zero T
	tracer := fctx.Tracer(reqCtx)
	return &ServiceContext[T]{
		Context:      ctx,
		ZapLogger:    fctx.Logger(reqCtx),
		Tracer:       &tracer,
		Props:        zero,
		RequestProps: map[string]interface{}{},
	}
}

// Deprecated: Use [fctx.Middleware] instead.
func ServiceContextMiddleware[T any](logger *zap.Logger, tracer *trace.Tracer, props T) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			req := ctx.Request()
			reqCtx := req.Context()

			var fields []zap.Field

			if spanCtx := trace.SpanContextFromContext(reqCtx); spanCtx.IsValid() {
				fields = append(fields,
					zap.String("trace_id", spanCtx.TraceID().String()),
					zap.String("span_id", spanCtx.SpanID().String()),
				)
			}

			if reqID := ctx.Response().Header().Get(echo.HeaderXRequestID); reqID != "" {
				fields = append(fields, zap.String("request_id", reqID))
				reqCtx = fctx.WithRequestID(reqCtx, reqID)
			}

			reqLogger := logger.With(fields...)

			reqCtx = fctx.WithLogger(reqCtx, reqLogger)
			if tracer != nil {
				reqCtx = fctx.WithTracer(reqCtx, *tracer)
			}

			ctx.SetRequest(req.WithContext(reqCtx))

			sctx := &ServiceContext[T]{
				Props:        props,
				RequestProps: map[string]interface{}{},
				Context:      ctx,
				ZapLogger:    reqLogger,
			}

			if tracer != nil {
				sctx.Tracer = tracer
			}

			return next(sctx)
		}
	}
}
