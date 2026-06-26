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

package telemetry

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ingka-group/fastecho/fctx"
)

// StartSpan starts a child span named after the calling function and returns
// the updated context and span. End the span with defer span.End().
//
// The span name is auto-discovered from the caller using runtime.Caller,
// formatted as package.Type.Method (e.g. "forecast.Service.Recompute").
//
// Uses the tracer from fctx.Tracer(ctx); falls back to a no-op tracer if none is set.
//
// For custom span names, use the standard OTel API directly:
//
//	tracer := otel.Tracer("my-scope")
//	ctx, span := tracer.Start(ctx, "custom-name")
func StartSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return fctx.Tracer(ctx).Start(ctx, callerName(1), opts...)
}

// SpanFunc wraps a function call in a span with the given name. The span starts
// before fn is called and ends after fn returns. Use this for tracing functions
// that don't accept context.Context:
//
//	var result T
//	telemetry.SpanFunc(ctx, "heavy-algorithm", func() {
//	    result = computeHeavyAlgorithm(data)
//	})
//
// If fn panics, the panic is recorded on the span and re-raised.
// If tracing is not configured, fn is still called (no-op span).
func SpanFunc(ctx context.Context, name string, fn func()) {
	_, span := fctx.Tracer(ctx).Start(ctx, name)
	defer span.End()
	defer func() {
		if r := recover(); r != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", r))
			span.RecordError(fmt.Errorf("panic: %v", r))
			panic(r)
		}
	}()
	fn()
}

// callerName returns the short function name of the caller at the given skip depth.
// Format: package.Type.Method or package.Function (strips module path and pointer receivers).
func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	name := runtime.FuncForPC(pc).Name()

	// Strip module path: "github.com/ingka-group/myservice/internal/forecast.(*Service).Recompute"
	// becomes "forecast.(*Service).Recompute"
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	// Strip pointer receiver: "forecast.(*Service).Recompute" becomes "forecast.Service.Recompute"
	name = strings.ReplaceAll(name, "(*", "")
	name = strings.ReplaceAll(name, ")", "")

	return name
}
