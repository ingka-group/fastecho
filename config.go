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
	ContextProps        any
	Opts                Opts
	Plugins             []Plugin
}

// Opts define configuration options for fastecho.
type Opts struct {
	Metrics      MetricsOpts
	Tracing      TracingOpts
	HealthChecks HealthChecksOpts
}

// MetricsOpts define configuration options for metrics.
type MetricsOpts struct {
	Skip bool
}

// TracingOpts define configuration options for tracing.
type TracingOpts struct {
	Skip        bool
	ServiceName string
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
