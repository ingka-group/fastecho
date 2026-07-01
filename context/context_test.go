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
