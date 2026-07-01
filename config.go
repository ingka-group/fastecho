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

package fastecho

import (
	"github.com/ingka-group/fastecho/env"
	"github.com/ingka-group/fastecho/router"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Config serves as input configuration for fastecho.
type Config struct {
	ExtraEnvs           env.Map
	ValidationRegistrar func(v *router.Validator) error
	Routes              func(e *echo.Echo, r *router.Router) error
	// ContextProps would be shared across all requests in the service
	ContextProps any
	Opts         Opts
	Plugins      []Plugin
	EchoFn       func(e *echo.Echo) error
	// Workers are long-running background processes managed by fastecho's
	// lifecycle, keyed by name. Each runs in its own goroutine with a context that
	// is cancelled on shutdown; its logs/spans are tagged worker=<name>.
	Workers map[string]Worker
}

// Opts define configuration options for fastecho.
type Opts struct {
	Metrics      MetricsOpts
	Tracing      TracingOpts
	Logs         LogsOpts
	HealthChecks HealthChecksOpts
}

// MetricsOpts define configuration options for metrics.
type MetricsOpts struct {
	Skip bool
}

// TracingOpts define configuration options for tracing.
type TracingOpts struct {
	Skip bool
}

// LogsOpts define configuration options for logging.
type LogsOpts struct {
	Skip bool // disable the access-log middleware entirely
}

// HealthChecksOpts define configuration options for health checks.
type HealthChecksOpts struct {
	Skip bool
	DB   *gorm.DB
}

type Plugin struct {
	ValidationRegistrar func(v *router.Validator) error
	Routes              func(e *echo.Echo, r *router.Router) error
}

func (c *Config) Use(p Plugin) {
	c.Plugins = append(c.Plugins, p)
}
