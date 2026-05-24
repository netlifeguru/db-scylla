package scylla

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

func parseConsistency(c string) gocql.Consistency {
	switch strings.ToUpper(c) {
	case "ONE":
		return gocql.One
	case "QUORUM":
		return gocql.Quorum
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum
	case "ALL":
		return gocql.All
	default:
		return gocql.LocalQuorum
	}
}

func contains(v string, list []string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func connectionFingerprint(cfg db.Config) string {
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}

	raw := fmt.Sprintf(
		"driver=%s host=%s port=%d keyspace=%s user=%s password=%s consistency=%s connect_timeout=%s",
		driverName,
		normalizeHost(cfg.Host),
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		strings.ToUpper(cfg.Consistency),
		connectTimeout.String(),
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func configFingerprint(cfg db.Config) string {
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}

	raw := fmt.Sprintf(
		"driver=%s host=%s port=%d keyspace=%s user=%s password=%s consistency=%s connect_timeout=%s",
		driverName,
		normalizeHost(cfg.Host),
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		strings.ToUpper(cfg.Consistency),
		connectTimeout.String(),
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func sessionFingerprint(cfg db.Config) string {
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}

	raw := fmt.Sprintf(
		"driver=%s host=%s port=%d keyspace=%s user=%s password=%s connect_timeout=%s",
		driverName,
		normalizeHost(cfg.Host),
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		connectTimeout.String(),
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (c *Connect) connectWithLimits(cfg db.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Host)

	if cfg.Port > 0 {
		cluster.Port = cfg.Port
	}

	if cfg.Database != "" {
		cluster.Keyspace = cfg.Database
	}

	if cfg.Username != "" || cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	cluster.ConnectTimeout = cfg.ConnectTimeout
	cluster.Timeout = c.timeout()
	cluster.PoolConfig.HostSelectionPolicy =
		gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.Consistency = parseConsistency(cfg.Consistency)
	cluster.NumConns = 4

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("create scylla session: %w", err)
	}

	return session, nil
}

func (c *Connect) CreatePool(cfg db.Config) error {
	cfg = normalizeConfig(cfg)

	if cfg.Host == "" {
		return db.ErrUnknownHost
	}

	if cfg.Identifier == "" {
		return db.ErrEmptyIdentifier
	}

	c.ensureShared()

	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if !contains(driverName, c.shared.myConnections) {
		c.shared.myConnections = append(c.shared.myConnections, driverName)
	}

	cfgKey := configFingerprint(cfg)
	sessKey := sessionFingerprint(cfg)

	if existing, ok := c.shared.Pools[cfg.Identifier]; ok {
		if existing.ConfigKey != cfgKey {
			return fmt.Errorf("%w: %q", db.ErrPoolIdentifierConflict, cfg.Identifier)
		}

		existing.Refs++

		c.Identifier = cfg.Identifier
		c.Host = cfg.Host
		return nil
	}

	var session *gocql.Session

	if existingShared, ok := c.shared.sharedPools[sessKey]; ok {
		existingShared.Refs++
		session = existingShared.Connection
	} else {
		created, err := c.connectWithLimits(cfg)
		if err != nil {
			return err
		}

		c.shared.sharedPools[sessKey] = &sharedPool{
			Connection: created,
			Refs:       1,
		}

		session = created
	}

	c.shared.Pools[cfg.Identifier] = &Pool{
		Connection:  session,
		Connect:     cfg,
		ConfigKey:   cfgKey,
		PoolKey:     sessKey,
		Refs:        1,
		Consistency: parseConsistency(cfg.Consistency),
	}

	c.Identifier = cfg.Identifier
	c.Host = cfg.Host

	return nil
}

func (c *Connect) Connection() *gocql.Session {
	p, err := c.currentPool()
	if err != nil {
		return nil
	}

	return p.Connection
}

func (c *Connect) Close() error {
	c.ensureShared()

	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if c.Identifier == "" {
		return nil
	}

	p, ok := c.shared.Pools[c.Identifier]
	if !ok {
		c.Identifier = ""
		return nil
	}

	if p.Refs > 1 {
		p.Refs--
		c.Identifier = ""
		return nil
	}

	delete(c.shared.Pools, c.Identifier)

	if shared, ok := c.shared.sharedPools[p.PoolKey]; ok {
		if shared.Refs > 1 {
			shared.Refs--
			c.Identifier = ""
			return nil
		}

		if shared.Connection != nil {
			shared.Connection.Close()
		}

		delete(c.shared.sharedPools, p.PoolKey)
	}

	c.Identifier = ""
	return nil
}

func (c *Connect) currentPool() (*Pool, error) {
	c.ensureShared()

	c.shared.poolsMu.RLock()
	defer c.shared.poolsMu.RUnlock()

	p, ok := c.shared.Pools[c.Identifier]
	if !ok || p == nil || p.Connection == nil {
		return nil, db.ErrNoConnection
	}

	return p, nil
}
