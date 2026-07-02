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
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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

	"github.com/ingka-group/fastecho/internal/banner"
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

// applyOTelEnvDefaults writes fastecho's OTEL_* compat defaults into the
// process environment, because that is the only channel the OTel SDK and
// autoexport read configuration from. Only unset vars are defaulted, and only
// for enabled signals, so operators keep full control and skipped telemetry
// leaves the environment untouched.
func applyOTelEnvDefaults(cfg Config) error {
	defaults := map[string]string{}
	if !cfg.SkipTraces {
		defaults[otelTracesExporter] = "otlp"
	}
	if !cfg.SkipMetrics {
		defaults[otelMetricsExporter] = "prometheus"
	}
	if !cfg.SkipTraces || !cfg.SkipMetrics {
		defaults[otelOTLPProtocol] = "grpc"
	}
	for name, value := range defaults {
		if os.Getenv(name) != "" {
			continue
		}
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("could not default %s: %w", name, err)
		}
	}
	return nil
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
}

// Shutdown flushes and stops all providers. Nil-safe.
func (p *Providers) Shutdown(ctx context.Context) error {
	if p == nil || p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

// Init builds the providers from OTEL_* env and returns them with one shutdown.
// It registers each provider (and, with tracing, the propagator) as the
// process-global OTel singleton — but only for enabled signals, so skipping a
// signal leaves an application's own OTel setup untouched.
// Call PrintConfiguration to log what it resolved.
func Init(ctx context.Context, cfg Config) (*Providers, error) {
	// Default the OTEL_* env so autoexport and the report read the same values.
	if err := applyOTelEnvDefaults(cfg); err != nil {
		return nil, err
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, err
	}

	// Warn but don't fail: unnamed services all collapse into "unknown_service".
	// The resource, not the env var, is checked: service.name may equally come
	// from OTEL_RESOURCE_ATTRIBUTES.
	if (!cfg.SkipTraces || !cfg.SkipMetrics) && !hasServiceName(res) {
		fmt.Fprintln(os.Stderr, "FastEcho telemetry: no service name is set; spans and metrics will report as \"unknown_service\", collapsing every unnamed service together. Set OTEL_SERVICE_NAME to identify this service.")
	}

	p := &Providers{
		TracerProvider: tracenoop.NewTracerProvider(),
		MeterProvider:  metricnoop.NewMeterProvider(),
	}
	var closers []func(context.Context) error

	// If a later step fails, stop whatever already started: the batch processor
	// goroutine and exporter connection have no other owner once Init returns nil.
	ok := false
	defer func() {
		if !ok {
			for _, c := range closers {
				_ = c(context.Background())
			}
		}
	}()

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
		if os.Getenv(otelMetricsExporter) == "prometheus" {
			reg := prometheus.NewRegistry()
			exp, err := promexporter.New(promexporter.WithRegisterer(reg))
			if err != nil {
				return nil, err
			}
			reader = exp
			// Serve the OTel metrics alongside everything on the default
			// registry (user-registered metrics, go/process collectors), so
			// /metrics keeps the content it had before the OTel migration.
			p.PrometheusGatherer = prometheus.Gatherers{reg, prometheus.DefaultGatherer}
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

	// Globals only for signals fastecho owns: with Skip set, a host app's own
	// providers/propagator must keep working.
	if !cfg.SkipTraces {
		otel.SetTracerProvider(p.TracerProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))
	}
	if !cfg.SkipMetrics {
		otel.SetMeterProvider(p.MeterProvider)
	}

	// Run the closers at most once: metric readers (unlike the trace provider)
	// return "reader is shutdown" if Shutdown is called twice, which would break
	// the nil-safe/idempotent contract callers rely on.
	var shutdownOnce sync.Once
	p.shutdown = func(ctx context.Context) error {
		var errs []error
		shutdownOnce.Do(func() {
			for _, c := range closers {
				errs = append(errs, c(ctx))
			}
		})
		return errors.Join(errs...)
	}

	p.cfg = cfg
	ok = true
	return p, nil
}

// PrintConfiguration prints the resolved telemetry configuration at startup.
// It reads the exported env (not stored state), so it is safe on any Providers
// value.
func (p *Providers) PrintConfiguration() {
	if p == nil {
		return
	}
	kvs := []string{
		"Service name", os.Getenv(otelServiceName),
		"Traces enabled", strconv.FormatBool(!p.cfg.SkipTraces),
	}
	if !p.cfg.SkipTraces {
		te := os.Getenv(otelTracesExporter)
		kvs = append(kvs, "Traces exporter", te)
		// OTLP transport is only meaningful when the traces exporter is OTLP.
		if te == "otlp" {
			proto := otlpProtocol(otelOTLPTracesProtocol)
			kvs = append(kvs,
				"Traces OTLP protocol", proto,
				"Traces OTLP endpoint", otlpEndpoint(otelOTLPTracesEndpoint, proto),
			)
		}
	}

	me := os.Getenv(otelMetricsExporter)
	delivery := "push"
	switch {
	case p.cfg.SkipMetrics:
		delivery = "off"
	case me == "prometheus":
		delivery = "pull (/metrics)"
	}
	kvs = append(kvs, "Metrics enabled", strconv.FormatBool(!p.cfg.SkipMetrics))
	if !p.cfg.SkipMetrics {
		kvs = append(kvs, "Metrics exporter", me)
	}
	kvs = append(kvs, "Metrics delivery", delivery)
	// OTLP transport is only meaningful when metrics push over OTLP.
	if !p.cfg.SkipMetrics && me == "otlp" {
		proto := otlpProtocol(otelOTLPMetricsProtocol)
		kvs = append(kvs,
			"Metrics OTLP protocol", proto,
			"Metrics OTLP endpoint", otlpEndpoint(otelOTLPMetricsEndpoint, proto),
		)
	}
	banner.Section("Telemetry configuration", kvs...)
}

// otlpProtocol resolves the OTLP transport for a signal: the signal-specific
// override wins, else the general protocol (defaulted by Init for enabled signals).
func otlpProtocol(signalKey string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	return os.Getenv(otelOTLPProtocol)
}

// otlpEndpoint reports the endpoint a signal will dial: the signal-specific var,
// then OTEL_EXPORTER_OTLP_ENDPOINT, else the SDK default for the protocol. With
// no env set the exporters dial localhost over TLS (host root CA) — an http://
// endpoint value is what enables plaintext, so the default must not print one.
func otlpEndpoint(signalKey, protocol string) string {
	if v := os.Getenv(signalKey); v != "" {
		return v
	}
	if v := os.Getenv(otelOTLPEndpoint); v != "" {
		return v
	}
	if protocol == "grpc" {
		return "localhost:4317 (default — TLS; set an http:// endpoint for plaintext)"
	}
	return "https://localhost:4318 (default)"
}

// newResource builds the shared resource: the SDK defaults (unknown_service
// fallback, telemetry.sdk.*) under a seeded service.instance.id, with env
// (OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES) winning every conflict.
func newResource(ctx context.Context) (*resource.Resource, error) {
	custom, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceInstanceID(uuid.New().String())),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, err
	}
	// custom is schemaless on purpose: Merge errors on two differing schema URLs.
	return resource.Merge(resource.Default(), custom)
}

// hasServiceName reports whether res carries a real service.name (not the
// SDK's unknown_service fallback).
func hasServiceName(res *resource.Resource) bool {
	v, ok := res.Set().Value(semconv.ServiceNameKey)
	return ok && v.AsString() != "" && !strings.HasPrefix(v.AsString(), "unknown_service")
}
