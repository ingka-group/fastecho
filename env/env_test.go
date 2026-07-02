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
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetEnv is a pure resolver: it fills the Map from the environment (or the
// declared defaults) and must never write back to the process environment —
// that would leak fastecho's defaults to other libraries and child processes.
func TestSetEnv_DoesNotMutateProcessEnv(t *testing.T) {
	const key = "FASTECHO_TEST_NO_EXPORT"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	m := Map{key: {DefaultValue: "resolved"}}
	require.NoError(t, m.SetEnv())

	assert.Equal(t, "resolved", m[key].Value, "default resolves into the Map")
	_, present := os.LookupEnv(key)
	assert.False(t, present, "SetEnv must not export resolved values to the process env")
}

// An explicit value wins over the default.
func TestSetEnv_ExplicitValueWinsOverDefault(t *testing.T) {
	const key = "FASTECHO_TEST_EXPLICIT"
	t.Setenv(key, "explicit")

	m := Map{key: {DefaultValue: "resolved"}}
	require.NoError(t, m.SetEnv())

	assert.Equal(t, "explicit", m[key].Value)
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

// A set-but-unknown LOG_LEVEL must not fail startup: it resolves as-is and the
// consumers normalize it via GetLogLevel (warn + dev), matching pre-OTel
// releases.
func TestSetEnv_AcceptsUnknownLogLevel(t *testing.T) {
	t.Setenv(LogLevel, "info")

	m := Map{LogLevel: NewLogLevelVar()}
	require.NoError(t, m.SetEnv(), "an unknown LOG_LEVEL must not fail startup")
	assert.Equal(t, "info", m[LogLevel].Value)
}

// The fallback must be visible: a set-but-unknown level prints a warning so
// the misconfiguration is not silent.
func TestGetLogLevel_WarnsOnUnknownValue(t *testing.T) {
	t.Setenv(LogLevel, "info")

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	level := GetLogLevel()
	os.Stdout = orig
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	assert.Equal(t, DevLogLevel, level)
	assert.Contains(t, string(out), "falling back", "the fallback warns so misconfiguration is visible")
}
