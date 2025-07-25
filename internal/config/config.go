package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/AvalosM/meli-interview-challenge/internal/cache"
	"github.com/AvalosM/meli-interview-challenge/internal/router"
	"github.com/AvalosM/meli-interview-challenge/internal/storage"
	"github.com/AvalosM/meli-interview-challenge/pkg/metrics"
	"github.com/AvalosM/meli-interview-challenge/pkg/shorturl"
)

// Config holds the configuration for the application
type Config struct {
	Logger          *LoggerConfig     `json:"logger"`
	Storage         *storage.Config   `json:"storage"`
	Cache           *cache.Config     `json:"cache"`
	ShortURLManager *shorturl.Config  `json:"short_url_manager"`
	MetricsManager  *metrics.Config   `json:"metrics_manager"`
	Router          *router.Config    `json:"router"`
	HTTPServer      *HTTPServerConfig `json:"http_server"`
}

type LoggerConfig struct {
	Level int `json:"level"`
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level: -4, // Default log level
	}
}

// Validate checks if the logger configuration is valid
func (c *LoggerConfig) Validate() error {
	switch slog.Level(c.Level) {
	case slog.LevelInfo, slog.LevelDebug, slog.LevelError, slog.LevelWarn:
		return nil
	default:
		return errors.New("invalid log level")
	}
}

// HTTPServerConfig holds the configuration for the HTTP server
type HTTPServerConfig struct {
	Port             int `json:"port"`
	ReadTimeoutInMS  int `json:"read_timeout_in_ms"`
	WriteTimeoutInMS int `json:"write_timeout_in_ms"`
	IdleTimeoutInMS  int `json:"idle_timeout_in_ms"`
}

// Validate checks if the HTTP server configuration is valid
func (c *HTTPServerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.ReadTimeoutInMS <= 0 {
		return fmt.Errorf("invalid read timeout: %d ms", c.ReadTimeoutInMS)
	}
	if c.WriteTimeoutInMS <= 0 {
		return fmt.Errorf("invalid write timeout: %d ms", c.WriteTimeoutInMS)
	}
	if c.IdleTimeoutInMS <= 0 {
		return fmt.Errorf("invalid idle timeout: %d ms", c.IdleTimeoutInMS)
	}

	return nil
}

// DefaultHTTPServerConfig returns a default HTTP server configuration
func DefaultHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port:             8080,
		ReadTimeoutInMS:  1000,
		WriteTimeoutInMS: 1000,
		IdleTimeoutInMS:  60000,
	}
}

// DefaultConfig returns a default configuration for the application
func DefaultConfig() *Config {
	return &Config{
		Logger:          DefaultLoggerConfig(),
		Storage:         storage.DefaultConfig(),
		Cache:           cache.DefaultConfig(),
		ShortURLManager: shorturl.DefaultConfig(),
		MetricsManager:  metrics.DefaultConfig(),
		Router:          router.DefaultConfig(),
		HTTPServer:      DefaultHTTPServerConfig(),
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := c.Logger.Validate(); err != nil {
		return err
	}
	if err := c.Storage.Validate(); err != nil {
		return err
	}
	if err := c.Cache.Validate(); err != nil {
		return err
	}
	if err := c.ShortURLManager.Validate(); err != nil {
		return err
	}
	if err := c.MetricsManager.Validate(); err != nil {
		return err
	}
	if err := c.Router.Validate(); err != nil {
		return err
	}
	if err := c.HTTPServer.Validate(); err != nil {
		return err
	}

	return nil
}
