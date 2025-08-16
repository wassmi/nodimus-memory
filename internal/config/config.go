package config

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds the application's configuration.
// It's populated from a TOML file.
//
// Note: The 'toml' struct tags are used to map the configuration file's
// keys to the struct fields.
type Config struct {
	Server  ServerConfig  `toml:"server"`
	Storage StorageConfig `toml:"storage"`
	Logger  LoggerConfig  `toml:"logger"`
}

// ServerConfig holds the server-related configuration.
type ServerConfig struct {
	Port    int    `toml:"port"`
	Bind    string `toml:"bind"`
	Timeout int    `toml:"timeout"`
}

// StorageConfig holds the storage-related configuration.
type StorageConfig struct {
	DataDir string `toml:"data_dir"`
}

// LoggerConfig holds the logger-related configuration.
type LoggerConfig struct {
	Level      string `toml:"level"`
	File       string `toml:"file"`
	MaxSize    int    `toml:"max_size"`
	MaxBackups int    `toml:"max_backups"`
	MaxAge     int    `toml:"max_age"`
	Compress   bool   `toml:"compress"`
}

// Load loads the configuration from the given file path.
func Load(path string) (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(path, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    4000,
			Bind:    "127.0.0.1",
			Timeout: 30,
		},
		Storage: StorageConfig{
			DataDir: "~/.nodimus",
		},
		Logger: LoggerConfig{
			Level:      "info",
			File:       "audit/nodimus.log",
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     30, // days
			Compress:   true,
		},
	}
}

// ExpandDataDir expands the tilde in the data directory path to the user's
// home directory.
func (c *Config) ExpandDataDir() (string, error) {
	if c.Storage.DataDir == "" {
		return "", nil
	}
	if strings.HasPrefix(c.Storage.DataDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(c.Storage.DataDir, "~", home, 1), nil
	}
	return os.ExpandEnv(c.Storage.DataDir), nil
}