package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/mikeunge/sshman/pkg/helpers"
)

const (
	defaultDatabasePath      = "~/.local/share/sshman/sshman.db"
	defaultLoggingPath       = "~/.local/share/sshman/sshman.log"
	defaultMaskInput         = true
	defaultDecryptionRetries = 1
)

type Config struct {
	DatabasePath      string `json:"databasepath"`
	LoggingPath       string `json:"logpath"`
	MaskInput         bool   `json:"maskInput"`
	DecryptionRetries int    `json:"decryptionRetries"`
}

// Paths to validate
var PathsToValidate = []string{
	"DatabasePath",
	"LoggingPath",
}

func Parse(path string) (Config, error) {
	var config = Config{}

	path = helpers.SanitizePath(path)
	if !helpers.FileExists(path) {
		return defaultConfig(), nil
	}

	data, err := helpers.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err = json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	config.sanitizeConfigPaths()
	if err := config.validatePaths(PathsToValidate, true); err != nil {
		return config, err
	}
	return config, nil
}

func (c *Config) validatePaths(objectNames []string, createIfNotExist bool) error {
	objectValues := reflect.ValueOf(*c)
	objectValueTypes := objectValues.Type()

	for i := 0; i < objectValues.NumField(); i++ {
		var (
			ok       bool
			objValue string
		)

		objName := objectValueTypes.Field(i).Name
		objValue = helpers.SanitizePath(objValue)

		if !slices.Contains(objectNames, objName) {
			continue
		}

		if objValue, ok = objectValues.Field(i).Interface().(string); !ok {
			return fmt.Errorf("Could not transform %+v into string.", objectValues.Field(i).Interface())
		}

		if helpers.PathExists(objValue) {
			continue
		} else if !helpers.PathExists(objValue) && !createIfNotExist {
			continue
		}

		if err := helpers.CreatePathIfNotExist(objValue); err != nil {
			return err
		}
	}
	return nil
}

func defaultConfig() Config {
	config := Config{
		DatabasePath:      defaultDatabasePath,
		LoggingPath:       defaultLoggingPath,
		MaskInput:         defaultMaskInput,
		DecryptionRetries: defaultDecryptionRetries,
	}
	config.sanitizeConfigPaths()
	config.validatePaths(PathsToValidate, true)
	return config
}

func (c *Config) sanitizeConfigPaths() {
	c.DatabasePath = helpers.SanitizePath(c.DatabasePath)
	c.LoggingPath = helpers.SanitizePath(c.LoggingPath)
}
