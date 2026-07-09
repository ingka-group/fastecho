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
	"fmt"
	"os"

	"github.com/ingka-group/fastecho/internal/stringutils"
)

const (
	// LogLevel is the environment variable name for the log level.
	LogLevel = "LOG_LEVEL"

	DevLogLevel  = "dev"
	TestLogLevel = "test"
	ProdLogLevel = "prod"
)

var logLevels = []string{DevLogLevel, TestLogLevel, ProdLogLevel}

// NewLogLevelVar returns the canonical LOG_LEVEL variable: defaults to dev.
// Deliberately not constrained via OneOf: an unknown value must not fail
// startup — consumers normalize it through GetLogLevel (warn + dev fallback),
// matching pre-OTel releases. Fresh *Var per call so Maps don't alias it.
func NewLogLevelVar() *Var {
	return &Var{
		DefaultValue: DevLogLevel,
	}
}

// GetLogLevel reads LOG_LEVEL, falling back to dev when unset or unknown.
// A set-but-unknown value warns so the misconfiguration is not silent.
func GetLogLevel() string {
	level := os.Getenv(LogLevel)
	if !stringutils.ExistsInSlice(level, logLevels) {
		if level != "" {
			fmt.Printf("no valid %s set (%q), falling back to %s\n", LogLevel, level, DevLogLevel)
		}
		return DevLogLevel
	}
	return level
}
