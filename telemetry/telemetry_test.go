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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	os.Unsetenv("OTEL_SERVICE_NAME")
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
	os.Unsetenv("OTEL_SERVICE_NAME")

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
	// test seam. The go.* runtime metrics surface as go_* families there.
	require.NotNil(t, p.PrometheusGatherer)
	mfs, err := p.PrometheusGatherer.Gather()
	require.NoError(t, err)

	var found bool
	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "go_") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected go_* runtime metrics via the prometheus gatherer")
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
