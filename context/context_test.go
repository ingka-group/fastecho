package context_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	fecontext "github.com/ingka-group/fastecho/context"
	"github.com/ingka-group/fastecho/fctx"
)

type compatFixture struct {
	ec     echo.Context
	logger *zap.Logger
	tracer trace.Tracer
}

func newCompatFixture(t *testing.T) compatFixture {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return compatFixture{
		ec:     e.NewContext(req, rec),
		logger: zaptest.NewLogger(t),
		tracer: noop.NewTracerProvider().Tracer("test"),
	}
}

func TestOldMiddleware_FctxAccessors(t *testing.T) {
	tests := map[string]struct {
		requestID string
		assert    func(t *testing.T, ctx context.Context)
	}{
		"logger accessible": {
			assert: func(t *testing.T, ctx context.Context) {
				got := fctx.Logger(ctx)
				require.NotNil(t, got)
				got.Info("safe to use")
			},
		},
		"tracer accessible": {
			assert: func(t *testing.T, ctx context.Context) {
				assert.NotNil(t, fctx.Tracer(ctx))
			},
		},
		"request ID propagated": {
			requestID: "old-mw-req-id",
			assert: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, "old-mw-req-id", fctx.RequestID(ctx))
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := newCompatFixture(t)
			if tc.requestID != "" {
				f.ec.Response().Header().Set(echo.HeaderXRequestID, tc.requestID)
			}

			handler := func(ec echo.Context) error {
				tc.assert(t, fctx.From(ec))
				return nil
			}

			err := fecontext.ServiceContextMiddleware[any](f.logger, &f.tracer, nil)(handler)(f.ec)
			require.NoError(t, err)
		})
	}
}

func TestOldMiddleware_NilTracer(t *testing.T) {
	f := newCompatFixture(t)

	handler := func(ec echo.Context) error {
		ctx := fctx.From(ec)
		assert.NotNil(t, fctx.Tracer(ctx), "returns noop fallback")

		sctx := fecontext.GetServiceContext[any](ec)
		assert.Nil(t, sctx.Tracer, "ServiceContext.Tracer stays nil")
		return nil
	}

	err := fecontext.ServiceContextMiddleware[any](f.logger, nil, nil)(handler)(f.ec)
	require.NoError(t, err)
}

func TestNewMiddleware_GetServiceContextShim(t *testing.T) {
	tests := map[string]struct {
		assert func(t *testing.T, sctx *fecontext.ServiceContext[any])
	}{
		"shim logger is not nil": {
			assert: func(t *testing.T, sctx *fecontext.ServiceContext[any]) {
				assert.NotNil(t, sctx.ZapLogger)
			},
		},
		"shim tracer pointer is not nil": {
			assert: func(t *testing.T, sctx *fecontext.ServiceContext[any]) {
				require.NotNil(t, sctx.Tracer)
				assert.NotNil(t, *sctx.Tracer)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := newCompatFixture(t)

			handler := func(ec echo.Context) error {
				tc.assert(t, fecontext.GetServiceContext[any](ec))
				return nil
			}

			err := fctx.Middleware(f.logger, f.tracer)(handler)(f.ec)
			require.NoError(t, err)
		})
	}
}

func TestOldMiddleware_GetServiceContext(t *testing.T) {
	f := newCompatFixture(t)

	handler := func(ec echo.Context) error {
		sctx := fecontext.GetServiceContext[any](ec)
		assert.NotNil(t, sctx.ZapLogger)
		assert.NotNil(t, sctx.Tracer)
		return nil
	}

	err := fecontext.ServiceContextMiddleware[any](f.logger, &f.tracer, nil)(handler)(f.ec)
	require.NoError(t, err)
}
