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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

// Without any name in the env, the resource must fall back to the SDK's
// unknown_service default (from resource.Default()) instead of omitting
// service.name entirely — that is what the startup warning promises.
func TestNewResource_FallsBackToUnknownService(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	res, err := newResource(t.Context())
	require.NoError(t, err)

	v, ok := res.Set().Value(semconv.ServiceNameKey)
	require.True(t, ok, "resource carries a service.name even when unconfigured")
	assert.True(t, strings.HasPrefix(v.AsString(), "unknown_service"),
		"unconfigured service.name falls back to unknown_service, got %q", v.AsString())
}

func TestNewResource_EnvServiceNameWins(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "named-svc")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	res, err := newResource(t.Context())
	require.NoError(t, err)

	v, ok := res.Set().Value(semconv.ServiceNameKey)
	require.True(t, ok)
	assert.Equal(t, "named-svc", v.AsString())
}

// The no-env default the exporters actually use is localhost with TLS, not
// plaintext http — the banner must not claim otherwise.
func TestOTLPEndpointDefaultIsAccurate(t *testing.T) {
	t.Setenv(otelOTLPTracesEndpoint, "")
	t.Setenv(otelOTLPEndpoint, "")

	grpcDefault := otlpEndpoint(otelOTLPTracesEndpoint, "grpc")
	assert.Contains(t, grpcDefault, "localhost:4317")
	assert.Contains(t, grpcDefault, "TLS")
	assert.NotContains(t, grpcDefault, "http://localhost", "grpc default must not present the endpoint as plaintext")

	httpDefault := otlpEndpoint(otelOTLPTracesEndpoint, "http/protobuf")
	assert.Contains(t, httpDefault, "https://localhost:4318")
}
