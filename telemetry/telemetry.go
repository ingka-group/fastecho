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

package telemetry

import (
	"context"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
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

	"github.com/ingka-group/fastecho/env"
)

// Config carries only wiring that has no env equivalent; behavior values
// (endpoint, sampling, exporter) come from OTEL_* env so there is one source of
// truth per setting.
type Config struct {
	SkipTraces  bool
	SkipMetrics bool

	// SetGlobal registers the providers + propagator as the process-global OTel
	// singletons (otelecho/otelhttp read the global propagator to carry
	// traceparent across hops). Tests pass false so parallel instances don't
	// clobber the global; not exposed on fastecho.Config because disabling it in
	// production would silently break propagation.
	SetGlobal bool
}

// Providers holds the configured providers and a single shutdown closer.
type Providers struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	// PrometheusGatherer is non-nil when OTEL_METRICS_EXPORTER=prometheus; the
	// router serves it at /metrics on the main port. Nil otherwise.
	PrometheusGatherer prometheus.Gatherer
	shutdown           func(context.Context) error
}

// Info is what Init resolved from OTEL_* env and the cfg toggles, for the caller
// to log once at startup.
type Info struct {
	ServiceName     string
	Traces          bool   // exporter active (not SkipTraces)
	TracesExporter  string // OTEL_TRACES_EXPORTER (default "otlp")
	OTLPProtocol    string // resolved transport: signal-specific var → general → "grpc"
	OTLPEndpoint    string // resolved: OTEL_EXPORTER_OTLP_ENDPOINT, else the protocol's SDK default
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
	// Preserve fastecho's historical gRPC default. autoexport reads
	// OTEL_EXPORTER_OTLP_PROTOCOL itself to pick the transport for BOTH the trace
	// and metric OTLP exporters, so setting the env var here is the single lever
	// that defaults every signal to gRPC - matching the old hardcoded behaviour and
	// keeping existing :4317 collectors working. Only when the operator expressed no
	// protocol preference; they opt into the OTel default by setting either var.
	if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "" &&
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL") == "" {
		_ = os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	}

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

		if err = otelruntime.Start(otelruntime.WithMeterProvider(mp)); err != nil {
			return nil, Info{}, err
		}
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

	return p, newInfo(cfg), nil
}

// newInfo snapshots what Init resolved from OTEL_* env and the cfg toggles, for
// the caller to log once at startup.
func newInfo(cfg Config) Info {
	info := Info{
		ServiceName:    os.Getenv("OTEL_SERVICE_NAME"),
		Traces:         !cfg.SkipTraces,
		TracesExporter: env.GetEnvVarOrDefault("OTEL_TRACES_EXPORTER", "otlp"),
		OTLPProtocol:   otlpProtocol("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"),
		OTLPEndpoint:   otlpEndpoint(),
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

	return info
}

// metricsExporter returns the configured metrics exporter, defaulting to
// prometheus so /metrics stays on the main port (the OTel SDK default is otlp).
func metricsExporter() string {
	return env.GetEnvVarOrDefault("OTEL_METRICS_EXPORTER", "prometheus")
}

// otlpProtocol resolves the OTLP transport: the signal-specific var wins, then
// OTEL_EXPORTER_OTLP_PROTOCOL. Defaults to grpc to match Init's backward-compat
// default; reading only the general var would misreport once an operator sets
// the signal-specific one.
func otlpProtocol(signalKey string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	return env.GetEnvVarOrDefault("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
}

// otlpEndpoint reports the OTLP endpoint that will be dialed: the operator's
// OTEL_EXPORTER_OTLP_ENDPOINT if set, else the SDK default for the resolved
// protocol (grpc => :4317, http/protobuf => :4318). Reporting the default
// rather than "(unset)" shows the address actually used.
func otlpEndpoint() string {
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return v
	}
	if otlpProtocol("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL") == "grpc" {
		return "http://localhost:4317 (default)"
	}
	return "http://localhost:4318 (default)"
}

// newResource builds the shared resource, seeding a default
// service.instance.id that env overrides on conflict (env is merged last).
func newResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(semconv.ServiceInstanceID(uuid.New().String())),
		resource.WithFromEnv(),
	)
}
