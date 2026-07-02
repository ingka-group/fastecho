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
	"fmt"
	"os"
	"strconv"
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

// OTEL_* environment variable names read by this package (and the OTel SDK).
const (
	otelServiceName         = "OTEL_SERVICE_NAME"
	otelTracesExporter      = "OTEL_TRACES_EXPORTER"
	otelMetricsExporter     = "OTEL_METRICS_EXPORTER"
	otelOTLPProtocol        = "OTEL_EXPORTER_OTLP_PROTOCOL"
	otelOTLPEndpoint        = "OTEL_EXPORTER_OTLP_ENDPOINT"
	otelOTLPTracesProtocol  = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"
	otelOTLPTracesEndpoint  = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	otelOTLPMetricsProtocol = "OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"
	otelOTLPMetricsEndpoint = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"
)

// otelEnv resolves fastecho's OTEL_* defaults into the environment
func otelEnv() env.Map {
	return env.Map{
		otelTracesExporter:  {DefaultValue: "otlp"},
		otelMetricsExporter: {DefaultValue: "prometheus"},
		otelOTLPProtocol:    {DefaultValue: "grpc"},
		otelServiceName:     {Optional: true},
	}
}

// Config carries only wiring that has no env equivalent; behavior values
// (endpoint, sampling, exporter) come from OTEL_* env so there is one source of
// truth per setting.
type Config struct {
	SkipTraces  bool
	SkipMetrics bool
}

// Providers holds the configured providers and a single shutdown closer.
type Providers struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	// PrometheusGatherer is non-nil when OTEL_METRICS_EXPORTER=prometheus; the
	// router serves it at /metrics on the main port. Nil otherwise.
	PrometheusGatherer prometheus.Gatherer

	shutdown func(context.Context) error
	cfg      Config
	env      env.Map
}

// Shutdown flushes and stops all providers. Nil-safe.
func (p *Providers) Shutdown(ctx context.Context) error {
	if p == nil || p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

// Init builds the providers from OTEL_* env and returns them with one shutdown.
// It registers the providers + propagator as the process-global OTel singletons.
// Call PrintConfiguration to log what it resolved.
func Init(ctx context.Context, cfg Config) (*Providers, error) {
	// Export the OTEL_* defaults so autoexport and the report read the same values.
	vars := otelEnv()
	if err := vars.SetEnv(); err != nil {
		return nil, err
	}

	// Warn but don't fail: without a name the SDK labels everything "unknown_service".
	if (!cfg.SkipTraces || !cfg.SkipMetrics) && vars[otelServiceName].Value == "" {
		fmt.Fprintln(os.Stderr, "FastEcho telemetry: OTEL_SERVICE_NAME is not set; spans and metrics will report as \"unknown_service\", collapsing every unnamed service together. Set OTEL_SERVICE_NAME to identify this service.")
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, err
	}

	p := &Providers{
		TracerProvider: tracenoop.NewTracerProvider(),
		MeterProvider:  metricnoop.NewMeterProvider(),
	}
	var closers []func(context.Context) error

	if !cfg.SkipTraces {
		exporter, err := autoexport.NewSpanExporter(ctx)
		if err != nil {
			return nil, err
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
		if vars[otelMetricsExporter].Value == "prometheus" {
			reg := prometheus.NewRegistry()
			exp, err := promexporter.New(promexporter.WithRegisterer(reg))
			if err != nil {
				return nil, err
			}
			reader = exp
			p.PrometheusGatherer = reg
		} else {
			r, err := autoexport.NewMetricReader(ctx)
			if err != nil {
				return nil, err
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
			return nil, err
		}
	}

	otel.SetTracerProvider(p.TracerProvider)
	otel.SetMeterProvider(p.MeterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

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

	p.cfg = cfg
	p.env = vars
	return p, nil
}

// PrintConfiguration prints the resolved telemetry configuration at startup.
func (p *Providers) PrintConfiguration() {
	if p == nil {
		return
	}
	fmt.Println("\nTelemetry configuration")
	printKV := func(k, v string) { fmt.Printf("  %-22s : %s\n", k, v) }

	tracesExporter := p.env[otelTracesExporter].Value
	printKV("Service name", p.env[otelServiceName].Value)
	printKV("Traces enabled", strconv.FormatBool(!p.cfg.SkipTraces))
	printKV("Traces exporter", tracesExporter)
	// OTLP transport is only meaningful when the traces exporter is OTLP.
	if !p.cfg.SkipTraces && tracesExporter == "otlp" {
		proto := otlpProtocol(p.env, otelOTLPTracesProtocol)
		printKV("Traces OTLP protocol", proto)
		printKV("Traces OTLP endpoint", otlpEndpoint(otelOTLPTracesEndpoint, proto))
	}

	me := p.env[otelMetricsExporter].Value
	delivery := "push"
	switch {
	case p.cfg.SkipMetrics:
		delivery = "off"
	case me == "prometheus":
		delivery = "pull (/metrics)"
	}
	printKV("Metrics enabled", strconv.FormatBool(!p.cfg.SkipMetrics))
	printKV("Metrics exporter", me)
	printKV("Metrics delivery", delivery)
	// OTLP transport is only meaningful when metrics push over OTLP.
	if !p.cfg.SkipMetrics && me == "otlp" {
		proto := otlpProtocol(p.env, otelOTLPMetricsProtocol)
		printKV("Metrics OTLP protocol", proto)
		printKV("Metrics OTLP endpoint", otlpEndpoint(otelOTLPMetricsEndpoint, proto))
	}
}

// otlpProtocol resolves the OTLP transport for a signal: the signal-specific
// override (SDK-only, not in the Map) wins, else the Map-resolved general protocol.
func otlpProtocol(vars env.Map, signalKey string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	return vars[otelOTLPProtocol].Value
}

// otlpEndpoint reports the endpoint a signal will dial: the signal-specific var,
// then OTEL_EXPORTER_OTLP_ENDPOINT, else the SDK default for the protocol. Read
// raw (no fastecho default) since the default is protocol-derived and must not
// be exported.
func otlpEndpoint(signalKey, protocol string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	if v := os.Getenv(otelOTLPEndpoint); v != "" {
		return v
	}
	if protocol == "grpc" {
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
