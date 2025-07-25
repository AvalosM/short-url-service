package storage

import "errors"

// Config contains the configuration for the storage connection
type Config struct {
	Driver         string `json:"driver"`
	DataSourceName string `json:"connection_string"`
}

// DefaultConfig returns the default configuration for the storage connection
func DefaultConfig() *Config {
	return &Config{
		Driver:         "pgx",
		DataSourceName: "postgres://postgres:postgres@localhost:5432/postgres",
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Driver == "" {
		return errors.New("driver cannot be empty")
	}
	if c.DataSourceName == "" {
		return errors.New("data source name cannot be empty")
	}
	return nil
}
