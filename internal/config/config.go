package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server
	Port string
	Env  string // development | staging | production

	// Logging
	LogLevel  string // debug | info | warn | error
	LogFormat string // text | json
	LogFile   string // file path, empty = stdout

	// Auth
	APIKeys []string // comma-separated in env

	// Rate limiting
	RateLimitRPS   int
	RateLimitBurst int

	// OpenTelemetry
	OTLPEndpoint string // empty = disabled
	ServiceName  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		LogFormat:      getEnv("LOG_FORMAT", "text"),
		LogFile:        getEnv("LOG_FILE", ""),
		OTLPEndpoint:   getEnv("OTLP_ENDPOINT", ""),
		ServiceName:    getEnv("SERVICE_NAME", "go-server-boilerplate"),
		RateLimitRPS:   getEnvInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
	}

	// Parse API keys
	raw := getEnv("API_KEYS", "")
	if raw != "" {
		for _, k := range strings.Split(raw, ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				cfg.APIKeys = append(cfg.APIKeys, k)
			}
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL %q: must be debug, info, warn, or error", c.LogLevel)
	}

	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.LogFormat] {
		return fmt.Errorf("invalid LOG_FORMAT %q: must be text or json", c.LogFormat)
	}

	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("RATE_LIMIT_RPS must be greater than 0")
	}

	if c.RateLimitBurst <= 0 {
		return fmt.Errorf("RATE_LIMIT_BURST must be greater than 0")
	}

	return nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// OTLPEnabled returns true when an OTLP endpoint is configured.
func (c *Config) OTLPEnabled() bool {
	return c.OTLPEndpoint != ""
}

// AuthEnabled returns true when at least one API key is configured.
func (c *Config) AuthEnabled() bool {
	return len(c.APIKeys) > 0
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}