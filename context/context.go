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

package context

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ServiceContext contains the echo.Context and custom properties vital for a microservice.
type ServiceContext[T any] struct {
	echo.Context
	ZapLogger *zap.Logger
	Tracer    *trace.Tracer
	// Props would be shared across all requests in the service
	Props T
	// RequestProps values are unique to each requests
	RequestProps map[string]interface{}
}

// BindValidate binds the data to the given interface and validates the input given using validator/10.
func (c *ServiceContext[T]) BindValidate(i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}

	if err := c.Validate(i); err != nil {
		return err
	}

	return nil
}

// GetServiceContext returns the ServiceContext from echo.Context.
func GetServiceContext[T any](ctx echo.Context) *ServiceContext[T] {
	return ctx.(*ServiceContext[T])
}

// ServiceContextMiddleware injects objects to echo.Context.
func ServiceContextMiddleware[T any](logger *zap.Logger, tracer *trace.Tracer, props T) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			req := ctx.Request()

			spanCtx := trace.SpanContextFromContext(req.Context())
			traceId := spanCtx.TraceID().String()
			spanId := spanCtx.SpanID().String()

			sctx := &ServiceContext[T]{
				Props:        props,
				RequestProps: map[string]interface{}{},
				Context:      ctx,
				ZapLogger: logger.With(
					zap.String("trace_id", traceId),
					zap.String("span_id", spanId),
				),
			}

			// Add the tracer to the service context
			if tracer != nil {
				sctx.Tracer = tracer
			}

			return next(sctx)
		}
	}
}
