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

package echozap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ingka-group/fastecho/env"
)

// ZapLevel maps the log level to a zapcore.Level
func ZapLevel(level string) zapcore.Level {
	switch level {
	case env.ProdLogLevel:
		return zapcore.WarnLevel
	case env.TestLogLevel:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

// New provides a logger with sane defaults for logging to server environments (dev, test, prod)
// It configures a JSON structured logger that writes info messages to stdout.
// The level is read from LOG_LEVEL, defaulting to dev when unset or unknown.
func New() (*zap.Logger, error) {
	level := env.GetLogLevel()

	var config zap.Config
	if level == env.ProdLogLevel {
		config = zap.NewProductionConfig()
	} else { // TestLogLevel, DevLogLevel
		config = zap.NewDevelopmentConfig()

		// Custom zap.NewDevelopmentConfig settings
		config.EncoderConfig = zap.NewProductionEncoderConfig()
		config.Encoding = "json" // Use structure logging
	}

	// Override log level based on fastecho logic above
	config.Level = zap.NewAtomicLevelAt(ZapLevel(level))

	// Use lowercase to prevent flakiness in log coloring in common tooling like Loki.
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// Use human-readable timestamp
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	// Make sure info level messages are written to stdout in all envs
	config.OutputPaths = []string{"stdout"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return zapLogger, nil
}
