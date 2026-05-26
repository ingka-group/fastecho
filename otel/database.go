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

package otel

import (
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// GormTracingOption is an alias for the upstream tracing option type.
// Re-exported so consumers don't need a direct dependency on
// gorm.io/plugin/opentelemetry/tracing.
type GormTracingOption = tracing.Option

// WithDBName returns a GormTracingOption that sets the db.namespace span attribute.
func WithDBName(name string) GormTracingOption {
	return tracing.WithAttributes(attribute.String("db.namespace", name))
}

// WithoutQueryVariables configures the plugin to exclude query variable values
// from the db.statement span attribute (PII safety).
var WithoutQueryVariables = tracing.WithoutQueryVariables

// UseGormTracing enables OpenTelemetry tracing on a GORM database connection.
//
// Options are passed directly to the upstream gorm.io/plugin/opentelemetry/tracing
// plugin. If the global TracerProvider is a noop (tracing disabled), spans are
// not produced — near-zero overhead.
//
// Usage:
//
//	otel.UseGormTracing(db)
//	otel.UseGormTracing(db, otel.WithDBName("bigquery"))
func UseGormTracing(db *gorm.DB, opts ...tracing.Option) error {
	return db.Use(tracing.NewPlugin(opts...))
}
