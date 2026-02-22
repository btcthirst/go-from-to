// Package config provides configuration management for the application.
package config

type Config struct {
	// Theme can be "light", "dark", or "system"

	theme      string
	ConfigPath string
}

func (c *Config) GetTheme() string {
	return c.theme
}

func Load() *Config {
	// For simplicity, we return a default config. In a real application, this would read from a file or environment variables.
	return &Config{
		theme: "system",
	}
}
