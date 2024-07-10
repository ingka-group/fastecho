package config

import (
	gocontext "context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// newTracer creates a new OTEL tracer.
func newTracer() (*sdktrace.TracerProvider, *oteltrace.Tracer, error) {
	if !Env[otelTracing].BooleanValue {
		return nil, nil, nil
	}

	exporter, err := otlptracehttp.New(gocontext.Background())
	if err != nil {
		return nil, nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	tracer := tp.Tracer(Env[ServiceName].Value)
	return tp, &tracer, nil
}
