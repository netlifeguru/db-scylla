package scylla

import (
	"strings"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/netlifeguru/db"
)

const defaultQueryTimeout = 5 * time.Second

type Pool struct {
	Connection  *gocql.Session
	Connect     db.Config
	ConfigKey   string
	PoolKey     string
	Refs        int
	Consistency gocql.Consistency
}

type sharedPool struct {
	Connection *gocql.Session
	Refs       int
}

type sharedState struct {
	poolsMu sync.RWMutex

	Pools         map[string]*Pool
	sharedPools   map[string]*sharedPool
	myConnections []string
}

type Connect struct {
	Host       string
	Identifier string
	Timeout    time.Duration
	shared     *sharedState
}

func New() *Connect {
	return &Connect{
		shared: newSharedState(),
	}
}

func newSharedState() *sharedState {
	return &sharedState{
		Pools:         make(map[string]*Pool),
		sharedPools:   make(map[string]*sharedPool),
		myConnections: make([]string, 0),
	}
}

func (c *Connect) ensureShared() {
	if c.shared == nil {
		c.shared = newSharedState()
	}
}

func normalizeHost(host string) string {
	return strings.Replace(host, "127.0.0.1", "localhost", 1)
}

func (c *Connect) LoadFile() string {
	return "model.cql"
}

func (c *Connect) DriverName() string {
	return "scylla"
}

func (c *Connect) Settings(identifier string) error {
	identifier = strings.TrimSpace(normalizeIdentifier(identifier))
	if identifier == "" {
		return db.ErrEmptyIdentifier
	}

	c.ensureShared()

	c.shared.poolsMu.RLock()
	defer c.shared.poolsMu.RUnlock()

	if _, ok := c.shared.Pools[identifier]; !ok {
		return db.ErrNoConnection
	}

	c.Identifier = identifier
	return nil
}

func (c *Connect) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultQueryTimeout
}

func (c *Connect) Fork() db.Conn {
	c.ensureShared()

	c.shared.poolsMu.Lock()
	defer c.shared.poolsMu.Unlock()

	if c.Identifier != "" {
		if p, ok := c.shared.Pools[c.Identifier]; ok && p != nil {
			p.Refs++
		}
	}

	return &Connect{
		Host:       c.Host,
		Identifier: c.Identifier,
		Timeout:    c.Timeout,
		shared:     c.shared,
	}
}
