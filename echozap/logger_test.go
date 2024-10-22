package echozap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		envType string
	}{
		{
			name:    "ok: log message NO env happy path",
			envType: "",
		},
		{
			name:    "ok: log message PROD env happy path",
			envType: ProdEnv,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(EnvType, test.envType)
			logger, err := New()

			assert.NoError(t, err)

			obs, logs := observer.New(zap.DebugLevel)
			logger = zap.New(zapcore.NewTee(logger.Core(), obs))

			message := "foo"
			logger.Info(message)
			assert.Equal(t, 1, logs.Len())

			logMessage := logs.AllUntimed()[0].Entry.Message
			assert.Equal(t, message, logMessage)
		})
	}
}
