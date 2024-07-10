package config

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
	// Hostname and the rest define ENV variables.
	Hostname        = "HOSTNAME"
	Port            = "PORT"
	EnvType         = "ENV_TYPE"
	SwaggerUITitle  = "SWAGGER_UI_TITLE"
	SwaggerJSONPath = "SWAGGER_JSON_PATH"
	ServiceName     = "SERVICE_NAME"

	otelTracing = "OTEL_TRACING"

	DBName            = "DB_NAME"
	DBHostname        = "DB_HOST"
	DBPort            = "DB_PORT"
	DBSSLMode         = "DB_SSL_MODE"
	DBUsername        = "DB_READ_WRITE_USER"
	DBPassword        = "DB_READ_WRITE_PASSWORD"
	DBMaxOpenConn     = "DB_MAX_OPEN_CONNECTIONS"
	DBMaxIdleConn     = "DB_MAX_IDLE_CONNECTIONS"
	DBMaxConnLifeTime = "DB_CONNECTION_MAX_LIFETIME"

	LocalEnv = "local"
	DevEnv   = "dev"
	TestEnv  = "test"
	ProdEnv  = "prod"
)

type EnvVar map[string]struct {
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
	Env = EnvVar{
		Hostname: {
			DefaultValue: "localhost",
		},
		Port: {
			DefaultValue: "8080",
			IsInteger:    true,
		},
		EnvType: {
			DefaultValue: DevEnv,
			OneOf:        []string{LocalEnv, DevEnv, TestEnv, ProdEnv},
		},
		SwaggerUITitle: {},
		SwaggerJSONPath: {
			// Defines the path to the swagger.json file on the server. This is used by the swagger UI.
			DefaultValue: "/swagger/swagger.json",
		},
		ServiceName: {},
		otelTracing: {
			DefaultValue: "false",
			IsBoolean:    true,
		},
	}

	DbEnv = EnvVar{
		DBHostname: {
			DefaultValue: "localhost",
		},
		DBPort: {
			DefaultValue: "5432",
			IsInteger:    true,
		},
		DBName:     {},
		DBUsername: {},
		DBPassword: {},
		DBSSLMode: {
			DefaultValue: "disable",
			OneOf:        []string{"enable", "disable"},
		},
		DBMaxOpenConn: {
			DefaultValue: "10",
			IsInteger:    true,
		},
		DBMaxIdleConn: {
			DefaultValue: "10",
			IsInteger:    true,
		},
		DBMaxConnLifeTime: {
			DefaultValue: "1h",
		},
	}
)

// loadEnvFile loads the variables of an env file. If it's not present, this step is skipped.
func loadEnvFile(filename string) {
	err := godotenv.Load(filename)
	// Don't stop or fail if the .env file doesn't exist.
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("'%s' configuration file doesn't exist. Don't worry, service will read the environment variables.", filename)
	}
}

// setEnv reads all the environment variables that are defined in the map and sets their values to the struct.
func setEnv(extraEnvVars *EnvVar, withPostgres bool) error {
	var messages []string

	// join extraEnv with Env
	for name, metadata := range *extraEnvVars {
		Env[name] = metadata
	}

	// conditionally join DB env variables
	if withPostgres {
		for name, metadata := range DbEnv {
			Env[name] = metadata
		}
	}

	loadEnvFile(".env")

	for name, metadata := range Env {
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
		Env[name] = metadata
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "; "))
	}

	return nil
}
