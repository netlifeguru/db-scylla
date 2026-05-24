package scylla

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

func newTestConnect() *Connect {
	return &Connect{
		Timeout: time.Second,
		shared:  newSharedState(),
	}
}

func testConfig(identifier string) db.Config {
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

func seedTestPool(c *Connect, cfg db.Config, refs int, sharedRefs int, withConnection bool) *Pool {
	c.ensureShared()

	cfg = normalizeConfig(cfg)

	cfgKey := configFingerprint(cfg)
	sessKey := sessionFingerprint(cfg)

	var session *gocql.Session
	if withConnection {
		session = &gocql.Session{}
	}

	p := &Pool{
		Connection:  session,
		Connect:     cfg,
		ConfigKey:   cfgKey,
		PoolKey:     sessKey,
		Refs:        refs,
		Consistency: parseConsistency(cfg.Consistency),
	}

	c.shared.Pools[cfg.Identifier] = p
	c.shared.sharedPools[sessKey] = &sharedPool{
		Connection: session,
		Refs:       sharedRefs,
	}

	c.Identifier = cfg.Identifier
	c.Host = cfg.Host

	return p
}
