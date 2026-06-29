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

package env

import (
	"os"
)

const (
	// LogLevel is the environment variable name for the log level.
	LogLevel = "LOG_LEVEL"

	DevLogLevel  = "dev"
	TestLogLevel = "test"
	ProdLogLevel = "prod"
)

// GetLogLevel reads LOG_LEVEL from the environment, falling back to DevLogLevel
// when unset or invalid.
func GetLogLevel() string {
	level := os.Getenv(LogLevel)
	if level != DevLogLevel && level != TestLogLevel && level != ProdLogLevel {
		return DevLogLevel
	}
	return level
}
