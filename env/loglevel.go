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

const (
	// LogLevel is the environment variable name for the log level.
	LogLevel = "LOG_LEVEL"

	DevLogLevel  = "dev"
	TestLogLevel = "test"
	ProdLogLevel = "prod"
)

// NewLogLevelVar returns the canonical LOG_LEVEL variable: defaults to dev,
// constrained to the known levels. Fresh *Var per call so Maps don't alias it.
func NewLogLevelVar() *Var {
	return &Var{
		DefaultValue: DevLogLevel,
		OneOf:        []string{DevLogLevel, TestLogLevel, ProdLogLevel},
	}
}
