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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

func TestNewResource_HonorsOperatorServiceInstanceID(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.instance.id=operator-set-id")
	res, err := newResource(t.Context())
	require.NoError(t, err)
	v, ok := res.Set().Value(semconv.ServiceInstanceIDKey)
	require.True(t, ok)
	assert.Equal(t, "operator-set-id", v.AsString())
}

func TestNewResource_GeneratesServiceInstanceIDWhenUnset(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")
	res, err := newResource(t.Context())
	require.NoError(t, err)
	v, ok := res.Set().Value(semconv.ServiceInstanceIDKey)
	require.True(t, ok)
	assert.NotEmpty(t, v.AsString())
}
