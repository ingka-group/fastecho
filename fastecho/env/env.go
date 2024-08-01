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

// Map is a map of environment variables.
// We use a *Var as map values are not addressable, i.e. cannot be modified directly.
type Map map[string]*Var

// Var describes an environment variable and its configuration.
type Var struct {
	Value        string
	DefaultValue string
	IsInteger    bool
	IntValue     int
	IsBoolean    bool
	BooleanValue bool
	OneOf        []string
	Optional     bool // controls whether an env variable can be missing from the .env file but still declared
}

// SetEnv reads and sets the provided list of env vars based on the Map.
func (m Map) SetEnv() error {
	var messages []string

	// Overwrite passed env var values via env file
	loadEnvFile(".env")

	for name, metadata := range m {
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
		m[name] = metadata
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "; "))
	}

	return nil
}

// loadEnvFile loads the variables of an env file. If it's not present, this step is skipped.
func loadEnvFile(filename string) {
	err := godotenv.Load(filename)
	// Don't stop or fail if the .env file doesn't exist.
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("'%s' configuration file doesn't exist. Don't worry, fastecho will read the environment variables.", filename)
	}
}
