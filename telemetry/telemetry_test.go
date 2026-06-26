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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ingka-group/fastecho/telemetry"
)

// initForTest builds providers for a unit test: the traces exporter kept off the
// network, the given config, and shutdown registered for cleanup. Tests set any
// extra OTEL_* envs (sampler, metrics exporter) before calling it.
func initForTest(t *testing.T, cfg telemetry.Config) *telemetry.Providers {
	t.Helper()
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	p, _, err := telemetry.Init(t.Context(), cfg) // discard the Info snapshot; these tests assert providers
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	return p
}

func TestInit_HonorsSamplerArg_RatioZeroDropsRoot(t *testing.T) {
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0") // sample nothing
	p := initForTest(t, telemetry.Config{SetGlobal: false})

	_, span := p.TracerProvider.Tracer("t").Start(t.Context(), "root")
	span.End()
	assert.False(t, span.SpanContext().IsSampled(), "ratio 0 => root span not sampled")
}

func TestInit_DefaultSamplingIsOn(t *testing.T) {
	// OTEL_TRACES_SAMPLER unset => SDK default parentbased_always_on (== 1.0).
	p := initForTest(t, telemetry.Config{SetGlobal: false})

	_, span := p.TracerProvider.Tracer("t").Start(t.Context(), "root")
	span.End()
	assert.True(t, span.SpanContext().IsSampled(), "default => sampled (today's behaviour)")
}

func TestInit_ResourceHasServiceName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "unit-svc")
	p := initForTest(t, telemetry.Config{SetGlobal: false})
	require.NotNil(t, p.TracerProvider)
}

func TestInit_TwoInstancesNoGlobal_DoNotInterfere(t *testing.T) {
	a := initForTest(t, telemetry.Config{SetGlobal: false})
	b := initForTest(t, telemetry.Config{SetGlobal: false})
	assert.NotSame(t, a.TracerProvider, b.TracerProvider)
}

// This test drives Shutdown directly, so it does its own Init. (initForTest
// registers a cleanup Shutdown, which would muddy the idempotency check.)
func TestInit_ShutdownIsNilSafeAndIdempotent(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	p, _, err := telemetry.Init(context.Background(), telemetry.Config{SetGlobal: false})
	require.NoError(t, err)

	require.NoError(t, p.Shutdown(context.Background()))
	require.NoError(t, p.Shutdown(context.Background()))

	var nilP *telemetry.Providers
	require.NoError(t, nilP.Shutdown(context.Background()))
}
