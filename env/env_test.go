// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetEnv must export resolved values (defaults included) back to the process
// environment, so os.Getenv-based readers - e.g. the OTel SDK - observe the
// same value the Map resolved.
func TestSetEnv_ExportsResolvedDefaultToProcessEnv(t *testing.T) {
	const key = "FASTECHO_TEST_EXPORT_DEFAULT"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	m := Map{key: {DefaultValue: "resolved"}}
	require.NoError(t, m.SetEnv())

	assert.Equal(t, "resolved", os.Getenv(key), "default must be exported to the environment")
	assert.Equal(t, "resolved", m[key].Value)
}

// An explicit value wins over the default and is exported unchanged.
func TestSetEnv_ExportsExplicitValue(t *testing.T) {
	const key = "FASTECHO_TEST_EXPORT_EXPLICIT"
	t.Setenv(key, "explicit")

	m := Map{key: {DefaultValue: "resolved"}}
	require.NoError(t, m.SetEnv())

	assert.Equal(t, "explicit", os.Getenv(key))
}

// An optional var with no value must stay unset rather than be exported empty,
// so the OTel SDK still treats it as absent.
func TestSetEnv_OptionalEmptyStaysUnset(t *testing.T) {
	const key = "FASTECHO_TEST_EXPORT_OPTIONAL"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	m := Map{key: {Optional: true}}
	require.NoError(t, m.SetEnv())

	_, present := os.LookupEnv(key)
	assert.False(t, present, "optional unset var must not be exported")
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name string
		set  string
		want string
	}{
		{name: "unset", want: DevLogLevel},
		{name: "invalid", set: "verbose", want: DevLogLevel},
		{name: "dev", set: DevLogLevel, want: DevLogLevel},
		{name: "test", set: TestLogLevel, want: TestLogLevel},
		{name: "prod", set: ProdLogLevel, want: ProdLogLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set == "" {
				os.Unsetenv(LogLevel)
				t.Cleanup(func() { os.Unsetenv(LogLevel) })
			} else {
				t.Setenv(LogLevel, tc.set)
			}

			assert.Equal(t, tc.want, GetLogLevel())
		})
	}
}
