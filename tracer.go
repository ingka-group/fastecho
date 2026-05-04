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
	gocontext "context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// newTracer creates a new OTEL tracer.
// Service name is read from OTEL_SERVICE_NAME env var via resource.WithFromEnv().
// Resource attributes (e.g. deployment.environment) can be set via OTEL_RESOURCE_ATTRIBUTES.
func newTracer() (*sdktrace.TracerProvider, *oteltrace.Tracer, error) {
	exporter, err := otlptracegrpc.New(gocontext.Background())
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.New(gocontext.Background(),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	tracer := tp.Tracer("github.com/ingka-group/fastecho")
	return tp, &tracer, nil
}
