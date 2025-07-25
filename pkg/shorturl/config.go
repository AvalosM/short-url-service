package shorturl

import "fmt"

// Config holds the configuration for the short URL manager
type Config struct {
	MaxShortURLIdRetries      int `json:"max_short_url_id_retries"`
	ShortURLCacheTTLInSeconds int `json:"short_url_cache_ttl_in_seconds"`
}

// DefaultConfig configuration
func DefaultConfig() *Config {
	return &Config{
		MaxShortURLIdRetries:      10,
		ShortURLCacheTTLInSeconds: 60 * 60, // 1 hour
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.MaxShortURLIdRetries <= 0 {
		return fmt.Errorf("MaxShortURLIdRetries must be greater than 0")
	}
	if c.ShortURLCacheTTLInSeconds <= 0 {
		return fmt.Errorf("ShortURLCacheTTLInSeconds must be greater than 0")
	}
	return nil
}
