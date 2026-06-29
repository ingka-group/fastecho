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

// ScopeName is the instrumentation scope name for fastecho's own spans.
const ScopeName = "github.com/ingka-group/fastecho/"

// StartSpan starts a child span named after the calling function (formatted as
// package.Type.Method, e.g. "forecast.Service.Recompute") and returns the updated
// context and span; end it with defer span.End(). It uses the tracer from
// fctx.Tracer(ctx), falling back to a no-op. For a custom name, call SpanFunc or
// the OTel tracer API directly.
func StartSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return fctx.Tracer(ctx).Start(ctx, callerName(1), opts...)
}

// SpanFunc runs fn inside a span named name, for tracing work that has no
// context.Context to thread. A panic in fn is recorded on the span and re-raised.
// With tracing off, fn still runs under a no-op span.
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

	// Strip the module path, leaving package.Type.Method.
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	// Strip pointer-receiver punctuation: "(*Service)" -> "Service".
	name = strings.ReplaceAll(name, "(*", "")
	name = strings.ReplaceAll(name, ")", "")

	return name
}
