package context_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap/zaptest"

	fecontext "github.com/ingka-group/fastecho/context"
	"github.com/ingka-group/fastecho/fctx"
)

func TestGetServiceContext_ShimUnderFctxMiddleware(t *testing.T) {
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
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ec := e.NewContext(req, rec)

			logger := zaptest.NewLogger(t)
			tracer := noop.NewTracerProvider().Tracer("test")

			handler := func(ec echo.Context) error {
				tc.assert(t, fecontext.GetServiceContext[any](ec))
				return nil
			}

			err := fctx.Middleware(logger, tracer)(handler)(ec)
			require.NoError(t, err)
		})
	}
}
