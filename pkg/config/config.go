package config

import (
	"encoding/json"

	"github.com/mikeunge/sshman/pkg/helpers"
)

type Config struct {
	DatabasePath string `json:"databasepath"`
	LoggingPath  string `json:"logpath"`
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

	// do some path sanitazation
	config.DatabasePath = helpers.SanitizePath(config.DatabasePath)
	config.LoggingPath = helpers.SanitizePath(config.LoggingPath)
	return config, nil
}

func defaultConfig() Config {
	return Config{
		DatabasePath: helpers.SanitizePath("~/.local/share/sshman.db"),
		LoggingPath:  helpers.SanitizePath("~/.local/share/sshman.log"),
	}
}
