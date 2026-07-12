package config

import sharedconfig "github.com/mashmool0/inama/libs/config"

type Config struct {
	ServiceName         string
	GRPCPort            int
	MetricsPort         int
	LogLevel            string
	DatabaseURL         string
	DefaultPageLimit    int32
	MaxPageLimit        int32
	AutoBootstrapSchema bool
}

func Load() Config {
	base := sharedconfig.LoadBase("user")
	defaultLimit, maxLimit := normalizePageLimits(
		int32(sharedconfig.GetInt("DEFAULT_PAGE_LIMIT", 20)),
		int32(sharedconfig.GetInt("MAX_PAGE_LIMIT", 100)),
	)

	return Config{
		ServiceName:         base.ServiceName,
		GRPCPort:            base.GRPCPort,
		MetricsPort:         base.MetricsPort,
		LogLevel:            base.LogLevel,
		DatabaseURL:         sharedconfig.MustGet("DATABASE_URL"),
		DefaultPageLimit:    defaultLimit,
		MaxPageLimit:        maxLimit,
		AutoBootstrapSchema: sharedconfig.GetBool("AUTO_BOOTSTRAP_SCHEMA", true),
	}
}

func normalizePageLimits(defaultLimit, maxLimit int32) (int32, int32) {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit < defaultLimit {
		maxLimit = defaultLimit
	}

	return defaultLimit, maxLimit
}
