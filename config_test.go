package scylla

import (
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func TestNormalizeIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "default"},
		{"spaces", "   ", "default"},
		{"trim and lowercase", "  Main  ", "main"},
		{"strip http", "http://Example.COM/", "example.com"},
		{"strip https", "https://Example.COM/", "example.com"},
		{"replace loopback", "127.0.0.1", "localhost"},
		{"replace loopback port", "127.0.0.1:9042", "localhost:9042"},
		{"keeps internal slash", "cluster/main", "cluster/main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeIdentifier(tt.in)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNormalizeConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg := normalizeConfig(db.Config{})

	if cfg.Identifier != defaultIdentifier {
		t.Fatalf("expected identifier %q, got %q", defaultIdentifier, cfg.Identifier)
	}

	if cfg.Host != defaultHost {
		t.Fatalf("expected host %q, got %q", defaultHost, cfg.Host)
	}

	if cfg.Port != defaultPort {
		t.Fatalf("expected port %d, got %d", defaultPort, cfg.Port)
	}

	if cfg.MaxConns != defaultMaxConns {
		t.Fatalf("expected max conns %d, got %d", defaultMaxConns, cfg.MaxConns)
	}

	if cfg.MinConns != defaultMinConns {
		t.Fatalf("expected min conns %d, got %d", defaultMinConns, cfg.MinConns)
	}

	if cfg.MaxConnIdleTime != defaultMaxConnIdleTime {
		t.Fatalf("expected max idle %s, got %s", defaultMaxConnIdleTime, cfg.MaxConnIdleTime)
	}

	if cfg.MaxConnLifetime != defaultMaxConnLifetime {
		t.Fatalf("expected max lifetime %s, got %s", defaultMaxConnLifetime, cfg.MaxConnLifetime)
	}

	if cfg.ConnectTimeout != defaultConnectTimeout {
		t.Fatalf("expected connect timeout %s, got %s", defaultConnectTimeout, cfg.ConnectTimeout)
	}

	if cfg.HealthCheckPeriod != defaultHealthCheckPeriod {
		t.Fatalf("expected health check period %s, got %s", defaultHealthCheckPeriod, cfg.HealthCheckPeriod)
	}

	if cfg.Consistency != defaultConsistency {
		t.Fatalf("expected consistency %q, got %q", defaultConsistency, cfg.Consistency)
	}
}

func TestNormalizeConfigPreservesProvidedValues(t *testing.T) {
	t.Parallel()

	cfg := normalizeConfig(db.Config{
		Identifier:        "Main",
		Host:              "db.internal",
		Port:              9142,
		MaxConns:          12,
		MinConns:          3,
		MaxConnIdleTime:   time.Minute,
		MaxConnLifetime:   2 * time.Hour,
		ConnectTimeout:    3 * time.Second,
		HealthCheckPeriod: 4 * time.Second,
		Consistency:       "ONE",
	})

	if cfg.Identifier != "main" {
		t.Fatalf("expected normalized identifier main, got %q", cfg.Identifier)
	}

	if cfg.Host != "db.internal" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}

	if cfg.Port != 9142 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}

	if cfg.MaxConns != 12 || cfg.MinConns != 3 {
		t.Fatalf("unexpected conn limits: max=%d min=%d", cfg.MaxConns, cfg.MinConns)
	}

	if cfg.MaxConnIdleTime != time.Minute {
		t.Fatalf("unexpected idle time: %s", cfg.MaxConnIdleTime)
	}

	if cfg.MaxConnLifetime != 2*time.Hour {
		t.Fatalf("unexpected lifetime: %s", cfg.MaxConnLifetime)
	}

	if cfg.ConnectTimeout != 3*time.Second {
		t.Fatalf("unexpected connect timeout: %s", cfg.ConnectTimeout)
	}

	if cfg.HealthCheckPeriod != 4*time.Second {
		t.Fatalf("unexpected health period: %s", cfg.HealthCheckPeriod)
	}

	if cfg.Consistency != "ONE" {
		t.Fatalf("unexpected consistency: %q", cfg.Consistency)
	}
}

func TestNormalizeConfigClampsMinConns(t *testing.T) {
	t.Parallel()

	cfg := normalizeConfig(db.Config{
		MaxConns: 2,
		MinConns: 10,
	})

	if cfg.MinConns != cfg.MaxConns {
		t.Fatalf("expected min conns to be clamped to max, got min=%d max=%d", cfg.MinConns, cfg.MaxConns)
	}
}

func TestNormalizeConfigNegativeMinConnsUsesDefault(t *testing.T) {
	t.Parallel()

	cfg := normalizeConfig(db.Config{MinConns: -1})

	if cfg.MinConns != defaultMinConns {
		t.Fatalf("expected default min conns, got %d", cfg.MinConns)
	}
}
