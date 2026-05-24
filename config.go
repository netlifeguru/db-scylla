package scylla

import (
	"strings"
	"time"

	"github.com/netlifeguru/db"
)

const driverName = "scylla"

const defaultIdentifier = "default"
const defaultHost = "127.0.0.1"
const defaultPort = 9042
const defaultMaxConns int32 = 25
const defaultMinConns int32 = 2
const defaultMaxConnIdleTime = 5 * time.Minute
const defaultMaxConnLifetime = time.Hour
const defaultConnectTimeout = 5 * time.Second
const defaultHealthCheckPeriod = 30 * time.Second
const defaultConsistency = "quorum"

func normalizeIdentifier(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return defaultIdentifier
	}

	identifier = strings.ToLower(identifier)
	identifier = strings.TrimPrefix(identifier, "http://")
	identifier = strings.TrimPrefix(identifier, "https://")
	identifier = strings.TrimSuffix(identifier, "/")

	if strings.HasPrefix(identifier, "127.0.0.1:") {
		identifier = "localhost:" + strings.TrimPrefix(identifier, "127.0.0.1:")
	}

	if identifier == "127.0.0.1" {
		identifier = "localhost"
	}

	return identifier
}

func normalizeConfig(cfg db.Config) db.Config {
	cfg.Identifier = normalizeIdentifier(cfg.Identifier)

	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = defaultHost
	}

	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}

	if cfg.MaxConns <= 0 {
		cfg.MaxConns = defaultMaxConns
	}

	if cfg.MinConns < 0 {
		cfg.MinConns = 0
	}

	if cfg.MinConns == 0 {
		cfg.MinConns = defaultMinConns
	}

	if cfg.MinConns > cfg.MaxConns {
		cfg.MinConns = cfg.MaxConns
	}

	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = defaultMaxConnIdleTime
	}

	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = defaultMaxConnLifetime
	}

	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = defaultConnectTimeout
	}

	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = defaultHealthCheckPeriod
	}

	if strings.TrimSpace(cfg.Consistency) == "" {
		cfg.Consistency = defaultConsistency
	}

	return cfg
}
