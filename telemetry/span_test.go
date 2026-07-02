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

package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/telemetry"
)

func TestStartSpan(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	tests := map[string]struct {
		ctx context.Context
	}{
		"creates span with context tracer": {
			ctx: ctx,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, span := telemetry.StartSpan(tc.ctx)
			defer span.End()

			require.NotNil(t, span)
			require.NotNil(t, ctx)
			assert.NotPanics(t, func() { span.End() })
		})
	}
}

func TestStartSpan_AutoDiscoversCallerName(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	_, span := telemetry.StartSpan(ctx)
	span.End()

	spans := exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Contains(t, spans[0].Name, "TestStartSpan_AutoDiscoversCallerName")
}

func TestStartSpan_ChildAttachesToParent(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	tracer := tp.Tracer("test")
	ctx, parent := tracer.Start(ctx, "parent")

	_, child := telemetry.StartSpan(ctx)
	child.End()
	parent.End()

	spans := exp.GetSpans()
	require.Len(t, spans, 2)

	childSpan := spans[0]
	parentSpan := spans[1]

	assert.Contains(t, childSpan.Name, "TestStartSpan_ChildAttachesToParent")
	assert.Equal(t, "parent", parentSpan.Name)
	assert.Equal(t, parentSpan.SpanContext.SpanID(), childSpan.Parent.SpanID())
}

func TestStartSpan_NoopWhenNoContextTracer(t *testing.T) {
	ctx, span := telemetry.StartSpan(t.Context())
	defer span.End()

	require.NotNil(t, ctx)
	require.NotNil(t, span)
	assert.NotPanics(t, func() { span.End() })
}

func TestStartSpan_UsesContextTracer(t *testing.T) {
	// A per-context in-memory provider, with NO global set.
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("ctx-scope"))

	_, span := telemetry.StartSpan(ctx)
	span.End()

	spans := exp.GetSpans()
	require.Len(t, spans, 1, "span recorded on the context provider, not the global")
}

func TestSpanFunc(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	var called bool
	telemetry.SpanFunc(ctx, "heavy-algorithm", func() {
		called = true
	})

	assert.True(t, called, "wrapped function should be called")

	spans := exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "heavy-algorithm", spans[0].Name)
}

func TestSpanFunc_ChildAttachesToParent(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	tracer := tp.Tracer("test")
	ctx, parent := tracer.Start(ctx, "parent")

	telemetry.SpanFunc(ctx, "child-work", func() {})
	parent.End()

	spans := exp.GetSpans()
	require.Len(t, spans, 2)
	assert.Equal(t, "child-work", spans[0].Name)
	assert.Equal(t, spans[1].SpanContext.SpanID(), spans[0].Parent.SpanID())
}

func TestSpanFunc_NoopWhenNoContextTracer(t *testing.T) {
	var called bool
	assert.NotPanics(t, func() {
		telemetry.SpanFunc(t.Context(), "noop-work", func() {
			called = true
		})
	})
	assert.True(t, called, "function should still execute even without tracing")
}

func TestSpanFunc_RecordsPanicAndReraises(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := fctx.WithTracer(t.Context(), tp.Tracer("test"))

	assert.PanicsWithValue(t, "something broke", func() {
		telemetry.SpanFunc(ctx, "panicking-work", func() {
			panic("something broke")
		})
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "panicking-work", spans[0].Name)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Contains(t, spans[0].Status.Description, "panic: something broke")
}

// A context not seeded by fastecho middleware (background jobs, cron
// callbacks) must still produce spans via the global provider when tracing is
// enabled — the pre-OTel-migration behavior.
func TestStartSpan_UnseededContextUsesGlobalProvider(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	otel.SetTracerProvider(tp)

	_, span := telemetry.StartSpan(context.Background())
	span.End()

	require.Len(t, exp.GetSpans(), 1, "unseeded context falls back to the global provider")
}

// The fctx-seeded tracer still takes priority over the global one.
func TestStartSpan_SeededTracerWinsOverGlobal(t *testing.T) {
	globalExp := tracetest.NewInMemoryExporter()
	globalTP := sdktrace.NewTracerProvider(sdktrace.WithSyncer(globalExp))
	t.Cleanup(func() { _ = globalTP.Shutdown(context.Background()) })
	prev := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	otel.SetTracerProvider(globalTP)

	seededExp := tracetest.NewInMemoryExporter()
	seededTP := sdktrace.NewTracerProvider(sdktrace.WithSyncer(seededExp))
	t.Cleanup(func() { _ = seededTP.Shutdown(context.Background()) })

	ctx := fctx.WithTracer(context.Background(), seededTP.Tracer("seeded"))
	_, span := telemetry.StartSpan(ctx)
	span.End()

	assert.Len(t, seededExp.GetSpans(), 1, "seeded tracer used")
	assert.Empty(t, globalExp.GetSpans(), "global provider not used when a tracer is seeded")
}
