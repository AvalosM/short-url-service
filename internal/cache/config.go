package cache

import "errors"

// Config holds the configuration for the cache connection.
type Config struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Protocol int    `json:"protocol"`
}

// DefaultConfig returns the default configuration for the cache connection.
func DefaultConfig() *Config {
	return &Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Addr == "" {
		return errors.New("addr cannot be empty")
	}
	if c.DB < 0 {
		return errors.New("db must be a non-negative integer")
	}
	if c.Protocol != 2 && c.Protocol != 3 {
		return errors.New("protocol must be either 2 or 3")
	}

	return nil
}
