// Package erebusconfig will contain the logic to parse the config toml file.
package erebusconfig

import (
	"github.com/BurntSushi/toml"
	"log/slog"
)

// Config contains the settings found inside the toml file.
type Config struct {
	StreamInterval float64
}

// Conf contains the setting.
var Conf Config
var confErr error

func init() {
	Conf, confErr = LoadConfig("config.toml")
	if confErr != nil {
		slog.Error("error parsing toml config file %w", "error", confErr.Error())
	}
}

// LoadConfig returns a Config.
func LoadConfig(path string) (Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return Config{}, err
	}
	Conf = c
	return Conf, nil
}
