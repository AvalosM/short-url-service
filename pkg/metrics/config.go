package metrics

import "errors"

// Config holds the configuration for the metrics manager
type Config struct {
	MetricsIntervalInMS      int `json:"metrics_interval_in_ms"`
	RequestChannelSize       int `json:"record_channel_size"`
	RecordRequestTimeoutInMS int `json:"record_request_timeout_in_ms"`
}

// DefaultConfig returns the default configuration for the metrics manager
func DefaultConfig() *Config {
	return &Config{
		MetricsIntervalInMS:      1000,
		RequestChannelSize:       1000,
		RecordRequestTimeoutInMS: 100,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.MetricsIntervalInMS <= 0 {
		return errors.New("MetricsIntervalInMS must be greater than 0")
	}
	if c.RequestChannelSize <= 0 {
		return errors.New("RequestChannelSize must be greater than 0")
	}
	if c.RecordRequestTimeoutInMS <= 0 {
		return errors.New("RecordRequestTimeoutInMS must be greater than 0")
	}
	return nil
}
