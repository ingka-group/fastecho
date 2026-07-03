// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
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

package echozap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ingka-group/fastecho/env"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{
			name:     "ok: log message NO env happy path",
			logLevel: "",
		},
		{
			name:     "ok: log message PROD level happy path",
			logLevel: env.ProdLogLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(env.LogLevel, test.logLevel)
			logger, err := New()

			assert.NoError(t, err)

			obs, logs := observer.New(zap.DebugLevel)
			logger = zap.New(zapcore.NewTee(logger.Core(), obs))

			message := "foo"
			logger.Info(message)
			assert.Equal(t, 1, logs.Len())

			logMessage := logs.AllUntimed()[0].Entry.Message
			assert.Equal(t, message, logMessage)
		})
	}
}
