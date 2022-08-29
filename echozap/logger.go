package echozap

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	EnvType = "ENV_TYPE"

	DevEnv  = "dev"
	TestEnv = "test"
	ProdEnv = "prod"
)

// getEnvType() get the env type from the OS env
func getEnvType() string {
	envType := os.Getenv(EnvType)

	if envType != DevEnv && envType != TestEnv && envType != ProdEnv {
		//Fallback with a warning
		fmt.Printf("no valid %s set, falling back to %s\n", EnvType, DevEnv)
		envType = DevEnv
	}

	return envType
}

// New provides a logger with sain defaults for logging to server ENVs (dev, test, prod)
// It configures a JSON structured logger that writes info messages to stdout
func New() (*zap.Logger, error) {
	envType := getEnvType()

	var config zap.Config
	if envType == ProdEnv {
		config = zap.NewProductionConfig()
	} else { // TestEnv, DevEnv
		config = zap.NewDevelopmentConfig()

		// Custom zap.NewDevelopmentConfig settings
		config.EncoderConfig = zap.NewProductionEncoderConfig()
		config.Encoding = "json" // Use structure logging
	}

	// Use CapitalLevelEncoder in all envs
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Use human-readable timestamp
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	// Make sure info level messages are written to stdout in all envs
	config.OutputPaths = []string{"stdout"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	defer func(zapLogger *zap.Logger) {
		_ = zapLogger.Sync()
	}(zapLogger)

	return zapLogger, nil
}
