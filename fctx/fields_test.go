package fctx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ingka-group/fastecho/fctx"
)

func TestFields_WithSpanAndRequestID(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("test").Start(t.Context(), "s")
	defer span.End()
	ctx = fctx.WithRequestID(ctx, "req-123")

	m := fieldMap(fctx.Fields(ctx))
	assert.Equal(t, span.SpanContext().TraceID().String(), m["trace_id"])
	assert.Equal(t, span.SpanContext().SpanID().String(), m["span_id"])
	assert.Equal(t, "req-123", m["request_id"])
}

func TestFields_NoSpan_NoTraceFields(t *testing.T) {
	m := fieldMap(fctx.Fields(t.Context()))
	assert.NotContains(t, m, "trace_id")
	assert.NotContains(t, m, "span_id")
	assert.NotContains(t, m, "request_id")
}

func TestNewRequestID_IsUUIDv4(t *testing.T) {
	id := fctx.NewRequestID()
	require.Len(t, id, 36)
	assert.Equal(t, byte('4'), id[14], "version nibble is 4")
}

// fieldMap logs the fields through an observer and reads them back as a map.
func fieldMap(fields []zap.Field) map[string]any {
	core, logs := observer.New(zapcore.DebugLevel)
	zap.New(core).Info("probe", fields...)
	return logs.All()[0].ContextMap()
}
