package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

const (
	Hostname        = "HOSTNAME"
	Port            = "PORT"
	EnvType         = "ENV_TYPE"
	SwaggerUITitle  = "SWAGGER_UI_TITLE"
	SwaggerJSONPath = "SWAGGER_JSON_PATH"
	ServiceName     = "SERVICE_NAME"

	OtelTracing = "OTEL_TRACING"

	localEnv = "local"
	devEnv   = "dev"
	testEnv  = "test"
	prodEnv  = "prod"
)

// EnvVars is a set of environment variables.
type EnvVars map[string]EnvVar

// EnvVar describes an environment variable and its configuration.
type EnvVar struct {
	Value        string
	DefaultValue string
	IsInteger    bool
	IntValue     int
	IsBoolean    bool
	BooleanValue bool
	OneOf        []string
	Optional     bool // controls whether an env variable can be missing from the .env file but still declared
}

var (
	DefaultEnvVars = EnvVars{
		Hostname: {
			DefaultValue: "localhost",
		},
		Port: {
			DefaultValue: "8080",
			IsInteger:    true,
		},
		EnvType: {
			DefaultValue: devEnv,
			OneOf:        []string{localEnv, devEnv, testEnv, prodEnv},
		},
		SwaggerUITitle: {},
		SwaggerJSONPath: {
			// Defines the path to the swagger.json file on the server. This is used by the swagger UI.
			DefaultValue: "/swagger/swagger.json",
		},
		ServiceName: {},
	}
)

// SetEnv reads and sets the provided list of env vars
func SetEnv(envVars EnvVars) (EnvVars, error) {
	var messages []string
	env := EnvVars{}

	// Overwrite passed env var values via env file
	loadEnvFile(".env")

	for name, metadata := range envVars {
		value := os.Getenv(name)
		if stringutils.IsEmpty(value) {
			if metadata.Optional || !stringutils.IsEmpty(metadata.DefaultValue) {
				value = metadata.DefaultValue
			} else {
				messages = append(messages,
					fmt.Sprintf("variable `%s` is required", name),
				)
			}
		}

		if len(metadata.OneOf) > 0 {
			if !stringutils.ExistsInSlice(value, metadata.OneOf) {
				messages = append(messages,
					fmt.Sprintf("variable `%s` must be set to one of %v", name, metadata.OneOf),
				)
			}
		}

		if metadata.IsInteger {
			intValue, err := stringutils.ToInt(value)
			if err != nil {
				messages = append(messages,
					fmt.Sprintf("variable `%s` requires an integer value", name),
				)
			}
			metadata.IntValue = intValue
		}

		if metadata.IsBoolean {
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				messages = append(messages,
					fmt.Sprintf("variable `%s` requires a boolean value", name),
				)
			}

			metadata.BooleanValue = boolValue
		}

		metadata.Value = value
		env[name] = metadata
	}

	if len(messages) > 0 {
		return nil, errors.New(strings.Join(messages, "; "))
	}

	return env, nil
}

// loadEnvFile loads the variables of an env file. If it's not present, this step is skipped.
func loadEnvFile(filename string) {
	err := godotenv.Load(filename)
	// Don't stop or fail if the .env file doesn't exist.
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("'%s' configuration file doesn't exist. Don't worry, service will read the environment variables.", filename)
	}
}
