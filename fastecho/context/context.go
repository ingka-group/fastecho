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
	Props     T
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
				Props:   props,
				Context: ctx,
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
