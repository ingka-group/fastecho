package otel

import (
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Option interface {
	apply(*TracerConfig)
}

type optionFunc func(*TracerConfig)

func (o optionFunc) apply(c *TracerConfig) {
	o(c)
}

type TracerConfig struct {
	TracerProvider oteltrace.TracerProvider
	Propagators    propagation.TextMapPropagator
	Skipper        middleware.Skipper
	ServiceName    string
	Env            string
}

func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *TracerConfig) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return optionFunc(func(cfg *TracerConfig) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithSkipper specifies a skipper for allowing requests to skip generating spans.
func WithSkipper(skipper middleware.Skipper) Option {
	return optionFunc(func(cfg *TracerConfig) {
		cfg.Skipper = skipper
	})
}

// WithServiceName specifies the name of the service
func WithServiceName(serviceName string) Option {
	return optionFunc(func(cfg *TracerConfig) {
		cfg.ServiceName = serviceName
	})
}

// WithEnv specifies the environment for filtering the spans
func WithEnv(env string) Option {
	return optionFunc(func(cfg *TracerConfig) {
		cfg.Env = env
	})
}
