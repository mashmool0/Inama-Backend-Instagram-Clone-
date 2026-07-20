// Package config loads configuration from environment variables into typed
// values. Twelve-factor style: config comes from the environment, never from
// checked-in files. Same helpers in every service.
package config

import (
	"log"
	"os"
	"strconv"
)

// Get returns the env var or a fallback default.
func Get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

// MustGet returns the env var or crashes the process. Use for values with no
// safe default (e.g. a database DSN) — fail fast at startup, not mid-request.
func MustGet(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		log.Fatalf("required env var %q is not set", key)
	}
	return v
}

// GetInt parses an int env var, falling back to def if unset or unparseable.
func GetInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// GetBool parses a bool env var (1/true/yes), falling back to def.
func GetBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// Base is the set of config every service shares. Load it, then read any
// service-specific vars on top.
type Base struct {
	ServiceName string
	GRPCPort    int    // port the gRPC server listens on
	MetricsPort int    // port /metrics + /health are served on
	LogLevel    string // debug|info|warn|error
}

// LoadBase reads the common config. Call once at startup.
func LoadBase(service string) Base {
	return Base{
		ServiceName: service,
		GRPCPort:    GetInt("GRPC_PORT", 50051),
		MetricsPort: GetInt("METRICS_PORT", 8080),
		LogLevel:    Get("LOG_LEVEL", "info"),
	}
}
