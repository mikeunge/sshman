package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/mikeunge/sshman/pkg/helpers"
)

const (
	defaultDatabasePath   = "~/.local/share/sshman/sshman.db"
	defaultLoggingPath    = "~/.local/share/sshman/sshman.log"
	defaultPrivateKeyPath = "~/.local/share/sshman/keys/"
)

type Config struct {
	DatabasePath   string `json:"databasepath"`
	LoggingPath    string `json:"logpath"`
	PrivateKeyPath string `json:"privateKeyPath"`
}

// Paths to validate
var PathsToValidate = []string{
	"DatabasePath",
	"LoggingPath",
	"PrivateKeyPath",
}

func Parse(path string) (Config, error) {
	var config = Config{}

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

	if err := config.validatePaths(PathsToValidate, true); err != nil {
		return config, err
	}
	return config, nil
}

func (c *Config) validatePaths(objectNames []string, createIfNotExist bool) error {
	objectValues := reflect.ValueOf(*c)
	objectValueTypes := objectValues.Type()

	for i := 0; i < objectValues.NumField(); i++ {
		var objValue string
		var ok bool

		if objValue, ok = objectValues.Field(i).Interface().(string); !ok {
			return fmt.Errorf("Could not transform %+v into string.", objectValues.Field(i).Interface())
		}
		objName := objectValueTypes.Field(i).Name
		objValue = helpers.SanitizePath(objValue)

		if !slices.Contains(objectNames, objName) {
			continue
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
		DatabasePath:   defaultDatabasePath,
		LoggingPath:    defaultLoggingPath,
		PrivateKeyPath: defaultPrivateKeyPath,
	}
	config.sanitizeConfigPaths()
	config.validatePaths(PathsToValidate, true)
	return config
}

func (c *Config) sanitizeConfigPaths() {
	c.DatabasePath = helpers.SanitizePath(c.DatabasePath)
	c.LoggingPath = helpers.SanitizePath(c.LoggingPath)
	c.PrivateKeyPath = helpers.SanitizePath(c.PrivateKeyPath)
}
