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

// telemetry.go is the OpenTelemetry bootstrap half of the package (span.go has
// StartSpan/SpanFunc + ScopeName). It builds one resource, env-driven exporters
// and sampling, and one shutdown closer shared by traces and metrics, so the two
// signals share identity and drain together.
package telemetry

import (
	"context"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// (ScopeName lives in span.go, the same package, so it is not redeclared here.)

// Config carries only wiring that has no env equivalent. Behaviour values
// (exporter endpoint, sampling, which exporter) come from OTEL_* env so there
// is one source of truth per setting; a struct field would duplicate that.
type Config struct {
	// SetGlobal registers the providers + propagator as the process-global OTel
	// singletons. Run/Initialize always pass true (otelecho/otelhttp read the
	// global propagator to carry traceparent across hops); tests pass false so
	// parallel instances don't clobber the global. Not exposed on fastecho.Config:
	// disabling it in production would silently break propagation.
	SetGlobal bool
	// SkipTraces / SkipMetrics disable a signal entirely.
	SkipTraces  bool
	SkipMetrics bool
}

// Providers holds the configured providers and a single shutdown closer. It is a
// pure runtime handle: the startup snapshot is returned separately from Init as an
// Info, not parked on the handle it has nothing to do with at runtime.
type Providers struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	// PrometheusGatherer is non-nil when OTEL_METRICS_EXPORTER=prometheus; the
	// router serves it at /metrics on the main port. Nil otherwise.
	PrometheusGatherer prometheus.Gatherer
	shutdown           func(context.Context) error
}

// Info is what Init resolved from OTEL_* env and the cfg toggles, for the caller to
// log once at startup (see fastecho config()). Init owns the env defaults, so it
// hands the resolution back rather than make the caller re-derive it.
type Info struct {
	ServiceName     string
	Traces          bool   // exporter active (not SkipTraces)
	TracesExporter  string // OTEL_TRACES_EXPORTER (default "otlp")
	OTLPProtocol    string // resolved transport (signal-specific var → general → "http/protobuf"); the fastecho server path forces "grpc" before Init (config()), so it reports "grpc" there but "http/protobuf" on a bare Init
	OTLPEndpoint    string // OTEL_EXPORTER_OTLP_ENDPOINT; "" => SDK default (grpc localhost:4317, http/protobuf localhost:4318/v1/<signal>)
	Metrics         bool
	MetricsExporter string // OTEL_METRICS_EXPORTER (default "prometheus")
	MetricsDelivery string // "pull (/metrics)" | "push" | "off"
}

// Shutdown flushes and stops all providers. Nil-safe.
func (p *Providers) Shutdown(ctx context.Context) error {
	if p == nil || p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

// Init builds the providers from OTEL_* env and returns them with one shutdown,
// plus an Info snapshot of what it resolved for the caller to log at startup.
func Init(ctx context.Context, cfg Config) (*Providers, Info, error) {
	res, err := newResource(ctx)
	if err != nil {
		return nil, Info{}, err
	}

	p := &Providers{
		TracerProvider: tracenoop.NewTracerProvider(),
		MeterProvider:  metricnoop.NewMeterProvider(),
	}
	var closers []func(context.Context) error

	if !cfg.SkipTraces {
		exporter, err := autoexport.NewSpanExporter(ctx)
		if err != nil {
			return nil, Info{}, err
		}
		// No WithSampler on purpose: the SDK reads OTEL_TRACES_SAMPLER / _ARG
		// itself, honouring every standard value and defaulting to
		// parentbased_always_on (100%). A hand-rolled sampler would only cover
		// a subset of those values and silently get the rest wrong.
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(exporter),
		)
		p.TracerProvider = tp
		closers = append(closers, tp.Shutdown)
	}

	if !cfg.SkipMetrics {
		var reader sdkmetric.Reader
		if metricsExporter() == "prometheus" {
			reg := prometheus.NewRegistry()
			exp, err := promexporter.New(promexporter.WithRegisterer(reg))
			if err != nil {
				return nil, Info{}, err
			}
			reader = exp
			p.PrometheusGatherer = reg
		} else {
			r, err := autoexport.NewMetricReader(ctx)
			if err != nil {
				return nil, Info{}, err
			}
			reader = r
		}
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(reader),
		)
		p.MeterProvider = mp
		closers = append(closers, mp.Shutdown)
	}

	if cfg.SetGlobal {
		otel.SetTracerProvider(p.TracerProvider)
		otel.SetMeterProvider(p.MeterProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))
	}

	// Run the closers at most once: metric readers (unlike the trace provider)
	// return "reader is shutdown" if Shutdown is called twice, which would break
	// the nil-safe/idempotent contract callers rely on.
	var shutdownOnce sync.Once
	p.shutdown = func(ctx context.Context) error {
		var firstErr error
		shutdownOnce.Do(func() {
			for _, c := range closers {
				if err := c(ctx); err != nil && firstErr == nil {
					firstErr = err
				}
			}
		})
		return firstErr
	}

	info := Info{
		ServiceName:    os.Getenv("OTEL_SERVICE_NAME"),
		Traces:         !cfg.SkipTraces,
		TracesExporter: envOr("OTEL_TRACES_EXPORTER", "otlp"),
		OTLPProtocol:   otlpProtocol("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"),
		OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}

	me := metricsExporter()
	info.Metrics = !cfg.SkipMetrics
	info.MetricsExporter = me
	switch {
	case cfg.SkipMetrics:
		info.MetricsDelivery = "off"
	case me == "prometheus":
		info.MetricsDelivery = "pull (/metrics)"
	default:
		info.MetricsDelivery = "push"
	}

	return p, info, nil
}

// metricsExporter returns the configured metrics exporter, defaulting to
// prometheus so /metrics stays on the main port (the OTel SDK default is otlp).
func metricsExporter() string {
	return envOr("OTEL_METRICS_EXPORTER", "prometheus") // envOr added in L3
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// otlpProtocol resolves the OTLP transport exactly the way autoexport does
// (spans.go/metrics.go): the signal-specific var wins, then the general
// OTEL_EXPORTER_OTLP_PROTOCOL, then the SDK default "http/protobuf". Reading only
// the general var would misreport (and mis-default) the moment an operator sets
// the signal-specific one - so Info would name a transport the exporter isn't using.
func otlpProtocol(signalKey string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	return envOr("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
}

// newResource builds the shared resource: service.name/version from env,
// service.instance.id generated if unset.
func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceInstanceID(instanceID()),
		),
	)
}

func instanceID() string {
	// resource.WithFromEnv honors OTEL_RESOURCE_ATTRIBUTES; if the operator set
	// service.instance.id there it wins on merge. Otherwise generate one.
	return uuid.New().String()
}
