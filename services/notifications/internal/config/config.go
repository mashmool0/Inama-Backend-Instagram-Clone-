package config

import (
	"strings"

	sharedconfig "github.com/mashmool0/inama/libs/config"
)

type Config struct {
	ServiceName         string
	GRPCPort            int
	MetricsPort         int
	LogLevel            string
	DatabaseURL         string
	DefaultPageLimit    int32
	MaxPageLimit        int32
	AutoBootstrapSchema bool
	RabbitMQURL         string
	RabbitMQExchange    string
	RabbitMQQueue       string
	RabbitMQRoutingKeys []string
}

func Load() Config {
	base := sharedconfig.LoadBase("notifications")
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
		RabbitMQURL:         sharedconfig.Get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		RabbitMQExchange:    sharedconfig.Get("RABBITMQ_EXCHANGE", "inama.events"),
		RabbitMQQueue:       sharedconfig.Get("RABBITMQ_QUEUE", "notif.q"),
		RabbitMQRoutingKeys: parseRoutingKeys(sharedconfig.Get("RABBITMQ_ROUTING_KEYS", "post.liked,comment.created,user.followed")),
	}
}

func parseRoutingKeys(raw string) []string {
	parts := strings.Split(raw, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			keys = append(keys, trimmed)
		}
	}
	return keys
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
