package router

// Config holds the configuration for the router
type Config struct {
	SwaggerEnabled bool `json:"swagger_enabled"`
}

// DefaultConfig returns the default configuration for the router
func DefaultConfig() *Config {
	return &Config{
		SwaggerEnabled: true, // Default to true for Swagger UI
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	return nil
}
