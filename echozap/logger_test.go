package echozap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLoggerMiddleware(logger)(h)(c)

	assert.Nil(t, err)

	logFields := logs.AllUntimed()[0].ContextMap()

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(200), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLoggerWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLoggerMiddlewareWithConfig(logger, ZapLoggerMiddlewareConfig{
		Skipper: func(ctx echo.Context) bool {
			return strings.Contains(ctx.Request().URL.Path, "/something")
		},
	})(h)(c)

	assert.Nil(t, err)

	assert.Equal(t, 0, logs.Len())
}

func TestNewServerLogger(t *testing.T) {
	tests := []struct {
		name        string
		envType     string
		expectError bool
	}{
		{
			name:        "Log message NON-PROD env happy path",
			envType:     DevEnv,
			expectError: false,
		},
		{
			name:        "Log message PROD env happy path",
			envType:     ProdEnv,
			expectError: false,
		},
		{
			name:        "Log message no env error path",
			envType:     "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(EnvTypeKey, test.envType)
			serverLogger, err := New()

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				obs, logs := observer.New(zap.DebugLevel)
				logger := zap.New(zapcore.NewTee(serverLogger.Core(), obs))

				message := "foo"
				logger.Info(message)
				assert.Equal(t, 1, logs.Len())
				logMessage := logs.AllUntimed()[0].Entry.Message
				assert.Equal(t, message, logMessage)
			}
		})
	}
}
