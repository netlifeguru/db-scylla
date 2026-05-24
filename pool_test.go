package scylla

import (
	"errors"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

func scyllaTestConfig(identifier string) db.Config {
	return db.Config{
		Identifier:     identifier,
		Host:           "localhost",
		Port:           9042,
		Database:       "app",
		Username:       "user",
		Password:       "secret",
		ConnectTimeout: 5 * time.Second,
		Consistency:    "LOCAL_QUORUM",
	}
}

func TestParseConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want gocql.Consistency
	}{
		{"one", "ONE", gocql.One},
		{"one lowercase", "one", gocql.One},
		{"quorum", "QUORUM", gocql.Quorum},
		{"local quorum", "LOCAL_QUORUM", gocql.LocalQuorum},
		{"all", "ALL", gocql.All},
		{"empty default", "", gocql.LocalQuorum},
		{"unknown default", "EACH_QUORUM", gocql.LocalQuorum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseConsistency(tt.in)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	if !contains("scylla", []string{"mysql", "scylla"}) {
		t.Fatalf("expected slice to contain scylla")
	}

	if contains("postgres", []string{"mysql", "scylla"}) {
		t.Fatalf("expected slice not to contain postgres")
	}

	if contains("scylla", nil) {
		t.Fatalf("expected nil slice not to contain scylla")
	}
}

func TestFingerprintsStable(t *testing.T) {
	t.Parallel()

	cfg := normalizeConfig(testConfig("main"))

	if connectionFingerprint(cfg) == "" {
		t.Fatalf("expected connection fingerprint")
	}

	if connectionFingerprint(cfg) != connectionFingerprint(cfg) {
		t.Fatalf("expected stable connection fingerprint")
	}

	if configFingerprint(cfg) != configFingerprint(cfg) {
		t.Fatalf("expected stable config fingerprint")
	}

	if sessionFingerprint(cfg) != sessionFingerprint(cfg) {
		t.Fatalf("expected stable session fingerprint")
	}
}

func TestFingerprintsNormalizeHostAndIgnoreIdentifier(t *testing.T) {
	t.Parallel()

	cfg1 := normalizeConfig(testConfig("main"))
	cfg1.Host = "localhost"

	cfg2 := normalizeConfig(testConfig("readonly"))
	cfg2.Host = "127.0.0.1"

	if configFingerprint(cfg1) != configFingerprint(cfg2) {
		t.Fatalf("expected config fingerprint to normalize host and ignore identifier")
	}

	if sessionFingerprint(cfg1) != sessionFingerprint(cfg2) {
		t.Fatalf("expected session fingerprint to normalize host and ignore identifier")
	}
}

func TestConfigFingerprintIncludesConsistency(t *testing.T) {
	t.Parallel()

	cfg1 := normalizeConfig(testConfig("main"))
	cfg1.Consistency = "ONE"

	cfg2 := normalizeConfig(testConfig("main"))
	cfg2.Consistency = "QUORUM"

	if configFingerprint(cfg1) == configFingerprint(cfg2) {
		t.Fatalf("expected config fingerprint to include consistency")
	}

	if sessionFingerprint(cfg1) != sessionFingerprint(cfg2) {
		t.Fatalf("expected session fingerprint to ignore consistency")
	}
}

func TestConfigFingerprintChangesWhenConfigChanges(t *testing.T) {
	t.Parallel()

	base := scyllaTestConfig("main")
	baseFP := configFingerprint(base)

	tests := []struct {
		name   string
		modify func(*db.Config)
	}{
		{"host", func(c *db.Config) { c.Host = "db.internal" }},
		{"port", func(c *db.Config) { c.Port = 9043 }},
		{"database", func(c *db.Config) { c.Database = "other" }},
		{"username", func(c *db.Config) { c.Username = "other" }},
		{"password", func(c *db.Config) { c.Password = "other" }},
		{"connect timeout", func(c *db.Config) { c.ConnectTimeout = 10 * time.Second }},
		{"consistency", func(c *db.Config) { c.Consistency = "ONE" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := base
			tt.modify(&cfg)

			got := configFingerprint(cfg)
			if got == baseFP {
				t.Fatalf("expected config fingerprint to change for %s", tt.name)
			}
		})
	}
}

func TestSessionFingerprintChangesWhenSessionConfigChanges(t *testing.T) {
	t.Parallel()

	base := scyllaTestConfig("main")
	baseFP := sessionFingerprint(base)

	tests := []struct {
		name   string
		modify func(*db.Config)
	}{
		{"host", func(c *db.Config) { c.Host = "db.internal" }},
		{"port", func(c *db.Config) { c.Port = 9043 }},
		{"database", func(c *db.Config) { c.Database = "other" }},
		{"username", func(c *db.Config) { c.Username = "other" }},
		{"password", func(c *db.Config) { c.Password = "other" }},
		{"connect timeout", func(c *db.Config) { c.ConnectTimeout = 10 * time.Second }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := base
			tt.modify(&cfg)

			got := sessionFingerprint(cfg)
			if got == baseFP {
				t.Fatalf("expected session fingerprint to change for %s", tt.name)
			}
		})
	}
}

func TestCreatePoolSameIdentifierDifferentConfigReturnsConflictBeforeConnect(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	cfg1 := testConfig("main")
	cfg2 := testConfig("main")
	cfg2.Database = "other"

	seedTestPool(c, cfg1, 1, 1, false)

	err := c.CreatePool(cfg2)
	if !errors.Is(err, db.ErrPoolIdentifierConflict) {
		t.Fatalf("expected ErrPoolIdentifierConflict, got %v", err)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected original pool to remain")
	}

	if p.Connect.Database != "app" {
		t.Fatalf("expected original database app, got %q", p.Connect.Database)
	}

	if p.Refs != 1 {
		t.Fatalf("expected refs to remain 1, got %d", p.Refs)
	}
}

func TestCreatePoolSameIdentifierSameConfigIncrementsRefsWithoutConnect(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	cfg := testConfig("main")
	seedTestPool(c, cfg, 1, 1, false)

	err := c.CreatePool(cfg)
	if err != nil {
		t.Fatalf("CreatePool returned error: %v", err)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected pool main")
	}

	if p.Refs != 2 {
		t.Fatalf("expected pool refs 2, got %d", p.Refs)
	}

	shared := c.shared.sharedPools[p.PoolKey]
	if shared == nil {
		t.Fatalf("expected shared pool")
	}

	if shared.Refs != 1 {
		t.Fatalf("expected shared refs to remain 1, got %d", shared.Refs)
	}
}

func TestCreatePoolDifferentIdentifierExistingSharedSessionWithoutConnect(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	cfg1 := normalizeConfig(testConfig("main"))
	cfg2 := normalizeConfig(testConfig("readonly"))

	sessKey := sessionFingerprint(cfg1)
	c.shared.sharedPools[sessKey] = &sharedPool{Refs: 1}

	err := c.CreatePool(cfg2)
	if err != nil {
		t.Fatalf("CreatePool returned error: %v", err)
	}

	if len(c.shared.Pools) != 1 {
		t.Fatalf("expected 1 logical pool, got %d", len(c.shared.Pools))
	}

	p := c.shared.Pools["readonly"]
	if p == nil {
		t.Fatalf("expected readonly pool")
	}

	if p.PoolKey != sessKey {
		t.Fatalf("expected pool key %q, got %q", sessKey, p.PoolKey)
	}

	if c.shared.sharedPools[sessKey].Refs != 2 {
		t.Fatalf("expected shared refs 2, got %d", c.shared.sharedPools[sessKey].Refs)
	}

	if len(c.shared.myConnections) != 1 || c.shared.myConnections[0] != driverName {
		t.Fatalf("expected driver recorded once, got %#v", c.shared.myConnections)
	}
}

func TestConnectionNoPoolReturnsNil(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	if conn := c.Connection(); conn != nil {
		t.Fatalf("expected nil connection, got %#v", conn)
	}
}

func TestCurrentPoolNoPoolReturnsError(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	p, err := c.currentPool()
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if p != nil {
		t.Fatalf("expected nil pool, got %#v", p)
	}
}

func TestCurrentPoolNilConnectionReturnsError(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	p, err := c.currentPool()
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if p != nil {
		t.Fatalf("expected nil pool, got %#v", p)
	}
}

func TestConnectionReturnsCurrentPoolSession(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	p := seedTestPool(c, testConfig("main"), 1, 1, true)

	if got := c.Connection(); got != p.Connection {
		t.Fatalf("expected current session")
	}
}

func TestCloseEmptyIdentifierDoesNothing(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no pools, got %d", len(c.shared.Pools))
	}
}

func TestCloseUnknownIdentifierClearsIdentifier(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	c.Identifier = "missing"

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}
}

func TestCloseDecrementsLogicalPoolRefs(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 2, 1, false)

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}

	p := c.shared.Pools["main"]
	if p == nil {
		t.Fatalf("expected logical pool to remain")
	}

	if p.Refs != 1 {
		t.Fatalf("expected pool refs 1, got %d", p.Refs)
	}
}

func TestCloseRemovesLastLogicalPoolAndSharedPool(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if c.Identifier != "" {
		t.Fatalf("expected identifier to be cleared, got %q", c.Identifier)
	}

	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected no logical pools, got %d", len(c.shared.Pools))
	}

	if len(c.shared.sharedPools) != 0 {
		t.Fatalf("expected no shared pools, got %d", len(c.shared.sharedPools))
	}
}

func TestCloseSharedPoolWithMultipleLogicalPools(t *testing.T) {
	t.Parallel()

	c := newTestConnect()

	p1 := seedTestPool(c, testConfig("main"), 1, 2, false)

	cfg2 := normalizeConfig(testConfig("readonly"))
	p2 := &Pool{
		Connection:  nil,
		Connect:     cfg2,
		ConfigKey:   configFingerprint(cfg2),
		PoolKey:     p1.PoolKey,
		Refs:        1,
		Consistency: parseConsistency(cfg2.Consistency),
	}

	c.shared.Pools["readonly"] = p2
	c.Identifier = "main"

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if _, ok := c.shared.Pools["main"]; ok {
		t.Fatalf("expected main pool to be removed")
	}

	if _, ok := c.shared.Pools["readonly"]; !ok {
		t.Fatalf("expected readonly pool to remain")
	}

	shared := c.shared.sharedPools[p1.PoolKey]
	if shared == nil {
		t.Fatalf("expected shared pool")
	}

	if shared.Refs != 1 {
		t.Fatalf("expected shared refs 1, got %d", shared.Refs)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	seedTestPool(c, testConfig("main"), 1, 1, false)

	if err := c.Close(); err != nil {
		t.Fatalf("first Close returned error: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("second Close returned error: %v", err)
	}

	if c.Identifier != "" {
		t.Fatalf("expected identifier to stay empty, got %q", c.Identifier)
	}
}

func TestCloseMissingSharedPoolDoesNotPanic(t *testing.T) {
	t.Parallel()

	c := newTestConnect()
	p := seedTestPool(c, testConfig("main"), 1, 1, false)
	delete(c.shared.sharedPools, p.PoolKey)

	err := c.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if len(c.shared.Pools) != 0 {
		t.Fatalf("expected logical pool to be removed, got %d", len(c.shared.Pools))
	}
}
