package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Geo    GeoConfig    `yaml:"geo"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port     int    `yaml:"port"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// TLSEnabled reports whether both the certificate and key paths are configured.
// When enabled, the server terminates TLS itself instead of relying on an
// upstream proxy.
func (c ServerConfig) TLSEnabled() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// DBConfig holds SQLite database settings.
type DBConfig struct {
	Path string `yaml:"path"`
}

// GeoConfig holds MaxMind GeoIP settings.
type GeoConfig struct {
	MMDBPath  string `yaml:"mmdb_path"`
	CacheSize int    `yaml:"cache_size"`
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{Port: 8080},
		DB:     DBConfig{Path: "/data/proxi-pulse.db"},
		Geo: GeoConfig{
			MMDBPath:  "/data/GeoLite2-City.mmdb",
			CacheSize: 50000,
		},
	}
}

// Load reads configuration from YAML file, falling back to defaults.
// Config path is read from PROXI_PULSE_CONFIG env var, defaulting to config.yaml.
func Load() *Config {
	cfg := defaultConfig()
	path := os.Getenv("PROXI_PULSE_CONFIG")
	if path == "" {
		path = "config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("config file not found, using defaults", "path", path)
		return cfg
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		slog.Warn("invalid config, using defaults", "error", err)
	}
	return cfg
}

// Level returns the default slog level.
func (c *Config) Level() slog.Level {
	return slog.LevelInfo
}
