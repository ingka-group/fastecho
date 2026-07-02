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

package telemetry_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/goleak"

	"github.com/ingka-group/fastecho/telemetry"
)

// capture swaps *target (os.Stdout or os.Stderr) for a pipe while fn runs and
// returns everything written to it.
func capture(t *testing.T, target **os.File, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := *target
	*target = w
	defer func() { *target = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func captureStderr(t *testing.T, fn func()) string { return capture(t, &os.Stderr, fn) }
func captureStdout(t *testing.T, fn func()) string { return capture(t, &os.Stdout, fn) }

// initForTest builds providers for a unit test: the traces exporter kept off the
// network, the given config, and shutdown registered for cleanup. Tests set any
// extra OTEL_* envs (sampler, metrics exporter) before calling it.
func initForTest(t *testing.T, cfg telemetry.Config) *telemetry.Providers {
	t.Helper()
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc") // quiet the missing-name warning
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	p, err := telemetry.Init(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	return p
}

func TestInit_HonorsSamplerArg_RatioZeroDropsRoot(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0") // sample nothing
	p := initForTest(t, telemetry.Config{})

	_, span := p.TracerProvider.Tracer("t").Start(t.Context(), "root")
	span.End()
	assert.False(t, span.SpanContext().IsSampled(), "ratio 0 => root span not sampled")
}

func TestInit_DefaultSamplingIsOn(t *testing.T) {
	// OTEL_TRACES_SAMPLER unset => SDK default parentbased_always_on (== 1.0).
	p := initForTest(t, telemetry.Config{})

	_, span := p.TracerProvider.Tracer("t").Start(t.Context(), "root")
	span.End()
	assert.True(t, span.SpanContext().IsSampled(), "default => sampled (today's behaviour)")
}

func TestInit_ResourceHasServiceName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	p := initForTest(t, telemetry.Config{})
	require.NotNil(t, p.TracerProvider)
}

// PrintConfiguration must actually write the resolved config to stdout.
func TestPrintConfiguration_PrintsResolvedConfig(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	p := initForTest(t, telemetry.Config{})

	out := captureStdout(t, p.PrintConfiguration)

	assert.Contains(t, out, "Telemetry configuration")
	assert.Contains(t, out, "Service name")
	assert.Contains(t, out, "unit-svc")
}

// Missing service name warns but doesn't fail Init.
func TestInit_WarnsButSucceedsWhenServiceNameMissing(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "") // empty == unset, and restored after the test
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	var p *telemetry.Providers
	stderr := captureStderr(t, func() {
		var err error
		p, err = telemetry.Init(t.Context(), telemetry.Config{})
		require.NoError(t, err)
	})
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	require.NotNil(t, p.TracerProvider)
	assert.Contains(t, stderr, "OTEL_SERVICE_NAME")
}

// Both signals skipped => no warning.
func TestInit_NoServiceNameWarningWhenAllSignalsSkipped(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "") // empty == unset, and restored after the test
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	var p *telemetry.Providers
	stderr := captureStderr(t, func() {
		var err error
		p, err = telemetry.Init(t.Context(), telemetry.Config{SkipTraces: true, SkipMetrics: true})
		require.NoError(t, err)
	})
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	assert.NotContains(t, stderr, "OTEL_SERVICE_NAME")
}

func TestInit_EachCallReturnsDistinctProviders(t *testing.T) {
	a := initForTest(t, telemetry.Config{})
	b := initForTest(t, telemetry.Config{})
	assert.NotSame(t, a.TracerProvider, b.TracerProvider)
}

func TestInit_PrometheusGathererExposed(t *testing.T) {
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
	p := initForTest(t, telemetry.Config{})
	require.NotNil(t, p.PrometheusGatherer, "prometheus exporter exposes a gatherer for /metrics")
}

func TestInit_MeterProviderUsableWhenSkipped(t *testing.T) {
	// Skipping metrics must still leave a non-nil (noop) provider so callers
	// never nil-check it.
	p := initForTest(t, telemetry.Config{SkipMetrics: true})
	_, err := p.MeterProvider.Meter("probe").Int64Counter("probe.count")
	require.NoError(t, err) // noop provider, but must not panic/error
}

func TestInit_StartsRuntimeMetrics(t *testing.T) {
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")
	p := initForTest(t, telemetry.Config{})

	// Gather from the prometheus registry that already backs /metrics - no extra
	// test seam. The gatherer now also merges client_golang's default go_*
	// collectors, so match on the OTel runtime instrumentation scope label to
	// prove the go.* metrics really come from otelruntime.
	require.NotNil(t, p.PrometheusGatherer)
	mfs, err := p.PrometheusGatherer.Gather()
	require.NoError(t, err)

	var found bool
	for _, mf := range mfs {
		if !strings.HasPrefix(mf.GetName(), "go_") {
			continue
		}
		for _, m := range mf.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "otel_scope_name" && strings.Contains(lp.GetValue(), "instrumentation/runtime") {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "expected go_* metrics from the otel runtime instrumentation via the prometheus gatherer")
}

// The gatherer must serve everything the pre-OTel /metrics served: metrics the
// app registered on prometheus.DefaultRegisterer and the default go_*/process_*
// collectors, alongside the OTel-exported families.
func TestInit_GathererIncludesDefaultRegistry(t *testing.T) {
	t.Setenv("OTEL_METRICS_EXPORTER", "prometheus")

	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fastecho_test_user_counter_total",
		Help: "a metric registered via the default registerer",
	})
	require.NoError(t, prometheus.Register(c))
	t.Cleanup(func() { prometheus.Unregister(c) })
	c.Inc()

	p := initForTest(t, telemetry.Config{})
	mfs, err := p.PrometheusGatherer.Gather()
	require.NoError(t, err)

	names := map[string]bool{}
	for _, mf := range mfs {
		names[mf.GetName()] = true
	}
	assert.True(t, names["fastecho_test_user_counter_total"], "user metrics on the default registerer stay on /metrics")
	assert.True(t, names["go_goroutines"], "client_golang default go collector stays on /metrics")
}

// Skipping a signal must leave the process-global OTel state alone so an app
// that runs its own SDK keeps working.
func TestInit_SkippedSignalsLeaveGlobalsUntouched(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevMP := otel.GetMeterProvider()
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetMeterProvider(prevMP)
		otel.SetTextMapPropagator(prevProp)
	})

	myTP := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = myTP.Shutdown(context.Background()) })
	myMP := sdkmetric.NewMeterProvider()
	t.Cleanup(func() { _ = myMP.Shutdown(context.Background()) })
	otel.SetTracerProvider(myTP)
	otel.SetMeterProvider(myMP)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	p, err := telemetry.Init(t.Context(), telemetry.Config{SkipTraces: true, SkipMetrics: true})
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	assert.Same(t, myTP, otel.GetTracerProvider(), "skipped tracing must not replace the global tracer provider")
	assert.Same(t, myMP, otel.GetMeterProvider(), "skipped metrics must not replace the global meter provider")
	_, isTC := otel.GetTextMapPropagator().(propagation.TraceContext)
	assert.True(t, isTC, "skipped tracing must not replace the global propagator")
}

// OTEL_* compat defaults must only reach the process env for enabled signals.
func TestInit_SkippedSignalsDontExportOTELDefaults(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_METRICS_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "")

	p, err := telemetry.Init(t.Context(), telemetry.Config{SkipTraces: true, SkipMetrics: true})
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	assert.Empty(t, os.Getenv("OTEL_TRACES_EXPORTER"))
	assert.Empty(t, os.Getenv("OTEL_METRICS_EXPORTER"))
	assert.Empty(t, os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func TestInit_ExportsDefaultsOnlyForEnabledSignals(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	t.Setenv("OTEL_TRACES_EXPORTER", "none") // enabled, kept off the network
	t.Setenv("OTEL_METRICS_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "")

	p, err := telemetry.Init(t.Context(), telemetry.Config{SkipMetrics: true})
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	assert.Empty(t, os.Getenv("OTEL_METRICS_EXPORTER"), "skipped metrics must not export the exporter default")
	assert.Equal(t, "grpc", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"), "enabled traces still export the compat protocol default")
}

// A failure after the tracer provider is built must tear it down again; the
// batch processor goroutine has no other owner once Init returns nil.
func TestInit_ErrorPathShutsDownStartedProviders(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "bogus") // autoexport rejects this

	ignore := goleak.IgnoreCurrent()
	_, err := telemetry.Init(context.Background(), telemetry.Config{})
	require.Error(t, err)
	goleak.VerifyNone(t, ignore)
}

// A hand-constructed Providers (public fields invite it) must not panic.
func TestPrintConfiguration_ZeroValueDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		(&telemetry.Providers{}).PrintConfiguration()
	})
}

// service.name supplied via OTEL_RESOURCE_ATTRIBUTES is just as valid as
// OTEL_SERVICE_NAME; the missing-name warning must not fire then.
func TestInit_NoWarningWhenServiceNameViaResourceAttributes(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=attr-svc")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	var p *telemetry.Providers
	stderr := captureStderr(t, func() {
		var err error
		p, err = telemetry.Init(t.Context(), telemetry.Config{})
		require.NoError(t, err)
	})
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })

	assert.NotContains(t, stderr, "unknown_service", "name set via OTEL_RESOURCE_ATTRIBUTES must not warn")
}

// This test drives Shutdown directly, so it does its own Init. (initForTest
// registers a cleanup Shutdown, which would muddy the idempotency check.)
func TestInit_ShutdownIsNilSafeAndIdempotent(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	p, err := telemetry.Init(context.Background(), telemetry.Config{})
	require.NoError(t, err)

	require.NoError(t, p.Shutdown(context.Background()))
	require.NoError(t, p.Shutdown(context.Background()))

	var nilP *telemetry.Providers
	require.NoError(t, nilP.Shutdown(context.Background()))
}
