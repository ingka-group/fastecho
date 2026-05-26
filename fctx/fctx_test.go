package fctx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/ingka-group/fastecho/fctx"
)

func TestLogger(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := map[string]struct {
		ctx      context.Context
		wantSame *zap.Logger
	}{
		"returns stored logger": {
			ctx:      fctx.WithLogger(context.Background(), logger),
			wantSame: logger,
		},
		"returns nop when missing": {
			ctx: context.Background(),
		},
		"returns nop when nil stored": {
			ctx: fctx.WithLogger(context.Background(), nil),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := fctx.Logger(tc.ctx)
			require.NotNil(t, got)
			got.Info("safe to use")

			if tc.wantSame != nil {
				assert.Same(t, tc.wantSame, got)
			}
		})
	}
}

func TestMustLogger(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := map[string]struct {
		ctx         context.Context
		shouldPanic bool
	}{
		"returns stored logger": {
			ctx: fctx.WithLogger(context.Background(), logger),
		},
		"panics when missing": {
			ctx:         context.Background(),
			shouldPanic: true,
		},
		"panics when nil stored": {
			ctx:         fctx.WithLogger(context.Background(), nil),
			shouldPanic: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() { fctx.MustLogger(tc.ctx) })
				return
			}
			assert.Same(t, logger, fctx.MustLogger(tc.ctx))
		})
	}
}

func TestTracer(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")

	tests := map[string]struct {
		ctx      context.Context
		wantSame bool
	}{
		"returns stored tracer": {
			ctx:      fctx.WithTracer(context.Background(), tracer),
			wantSame: true,
		},
		"returns noop fallback when missing": {
			ctx: context.Background(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := fctx.Tracer(tc.ctx)
			require.NotNil(t, got)

			if tc.wantSame {
				assert.Equal(t, tracer, got)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	tests := map[string]struct {
		ctx  context.Context
		want string
	}{
		"returns ID when present": {
			ctx:  fctx.WithRequestID(context.Background(), "req-123"),
			want: "req-123",
		},
		"returns empty when missing": {
			ctx:  context.Background(),
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, fctx.RequestID(tc.ctx))
		})
	}
}
