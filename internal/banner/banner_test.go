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

package banner

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSectionPrintsTitleAndAlignedPairs(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	Section("Log configuration", "LOG_LEVEL (env)", "dev", "EchoZap level", "debug")

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	assert.Contains(t, string(out), "Log configuration")
	assert.Contains(t, string(out), "  LOG_LEVEL (env)        : dev")
	assert.Contains(t, string(out), "  EchoZap level          : debug")
}
