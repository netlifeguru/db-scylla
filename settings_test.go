package scylla

import (
	"errors"
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func TestNew(t *testing.T) {
	t.Parallel()

	c := New()
	if c == nil {
		t.Fatalf("expected connection")
	}

	if c.shared == nil {
		t.Fatalf("expected shared state")
	}

	if c.shared.Pools == nil {
		t.Fatalf("expected Pools map")
	}

	if c.shared.sharedPools == nil {
		t.Fatalf("expected sharedPools map")
	}

	if c.shared.myConnections == nil {
		t.Fatalf("expected myConnections slice")
	}
}

func TestNewSharedState(t *testing.T) {
	t.Parallel()

	s := newSharedState()
	if s == nil {
		t.Fatalf("expected shared state")
	}

	if s.Pools == nil {
		t.Fatalf("expected Pools map")
	}

	if s.sharedPools == nil {
		t.Fatalf("expected sharedPools map")
	}

	if s.myConnections == nil {
		t.Fatalf("expected myConnections slice")
	}
}

func TestEnsureShared(t *testing.T) {
	t.Parallel()

	c := &Connect{}
	if c.shared != nil {
		t.Fatalf("expected nil shared before ensureShared")
	}

	c.ensureShared()

	if c.shared == nil {
		t.Fatalf("expected shared after ensureShared")
	}

	if c.shared.Pools == nil || c.shared.sharedPools == nil {
		t.Fatalf("expected initialized shared maps")
	}
}

func TestEnsureSharedKeepsExistingSharedState(t *testing.T) {
	t.Parallel()

	s := newSharedState()
	c := &Connect{shared: s}

	c.ensureShared()

	if c.shared != s {
		t.Fatalf("expected existing shared state to be preserved")
	}
}

func TestNormalizeHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"localhost", "localhost", "localhost"},
		{"replace loopback", "127.0.0.1", "localhost"},
		{"replace first loopback only", "127.0.0.1,127.0.0.1", "localhost,127.0.0.1"},
		{"spaces preserved", " 127.0.0.1 ", " localhost "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeHost(tt.in)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	t.Parallel()

	c := New()

	if got := c.LoadFile(); got != "model.cql" {
		t.Fatalf("expected model.cql, got %q", got)
	}
}

func TestDriverName(t *testing.T) {
	t.Parallel()

	c := New()

	if got := c.DriverName(); got != "scylla" {
		t.Fatalf("expected scylla, got %q", got)
	}
}

func TestSettingsRejectsUnknownIdentifier(t *testing.T) {
	t.Parallel()

	c := New()

	err := c.Settings("missing")
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if c.Identifier != "" {
		t.Fatalf("expected identifier to remain empty, got %q", c.Identifier)
	}
}

func TestSettingsBlankIdentifierUsesDefault(t *testing.T) {
	t.Parallel()

	c := New()
	seedTestPool(c, testConfig("default"), 1, 1, false)
	c.Identifier = ""

	err := c.Settings("   ")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}

	if c.Identifier != "default" {
		t.Fatalf("expected default identifier, got %q", c.Identifier)
	}
}

func TestSettingsInitializesSharedState(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	err := c.Settings("main")
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if c.shared == nil {
		t.Fatalf("expected Settings to initialize shared state")
	}
}

func TestSettingsSwitchesIdentifier(t *testing.T) {
	t.Parallel()

	c := New()
	seedTestPool(c, testConfig("main"), 1, 1, false)
	c.Identifier = ""

	err := c.Settings("main")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}

	if c.Identifier != "main" {
		t.Fatalf("expected identifier main, got %q", c.Identifier)
	}
}

func TestSettingsTrimsAndNormalizesIdentifier(t *testing.T) {
	t.Parallel()

	c := New()
	seedTestPool(c, testConfig("localhost"), 1, 1, false)
	c.Identifier = ""

	err := c.Settings("  127.0.0.1  ")
	if err != nil {
		t.Fatalf("Settings returned error: %v", err)
	}

	if c.Identifier != "localhost" {
		t.Fatalf("expected identifier localhost, got %q", c.Identifier)
	}
}

func TestTimeoutDefault(t *testing.T) {
	t.Parallel()

	c := New()

	if got := c.timeout(); got != defaultQueryTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultQueryTimeout, got)
	}
}

func TestTimeoutCustom(t *testing.T) {
	t.Parallel()

	c := New()
	c.Timeout = 10 * time.Second

	if got := c.timeout(); got != 10*time.Second {
		t.Fatalf("expected custom timeout 10s, got %s", got)
	}
}

func TestForkInitializesSharedState(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	forked := c.Fork()
	if forked == nil {
		t.Fatalf("expected forked connection")
	}

	if c.shared == nil {
		t.Fatalf("expected Fork to initialize source shared state")
	}

	fc, ok := forked.(*Connect)
	if !ok {
		t.Fatalf("expected *Connect, got %T", forked)
	}

	if fc.shared != c.shared {
		t.Fatalf("expected shared state to be reused")
	}
}

func TestForkSharesStateAndCopiesConnectionSettings(t *testing.T) {
	t.Parallel()

	c := New()
	c.Host = "localhost"
	c.Identifier = "main"
	c.Timeout = 10 * time.Second

	p := seedTestPool(c, testConfig("main"), 1, 1, false)

	forked := c.Fork()
	if forked == nil {
		t.Fatalf("expected forked connection")
	}

	fc, ok := forked.(*Connect)
	if !ok {
		t.Fatalf("expected *Connect, got %T", forked)
	}

	if fc == c {
		t.Fatalf("expected different Connect instance")
	}

	if fc.shared != c.shared {
		t.Fatalf("expected shared state to be reused")
	}

	if fc.Host != c.Host || fc.Identifier != c.Identifier || fc.Timeout != c.Timeout {
		t.Fatalf("expected connection settings to be copied")
	}

	if p.Refs != 2 {
		t.Fatalf("expected refs 2 after fork, got %d", p.Refs)
	}
}

func TestForkWithoutIdentifierDoesNotIncrementRefs(t *testing.T) {
	t.Parallel()

	c := New()
	p := seedTestPool(c, testConfig("main"), 1, 1, false)
	c.Identifier = ""

	_ = c.Fork()

	if p.Refs != 1 {
		t.Fatalf("expected refs to remain 1, got %d", p.Refs)
	}
}

func TestForkUnknownIdentifierDoesNotPanic(t *testing.T) {
	t.Parallel()

	c := New()
	c.Identifier = "missing"

	forked := c.Fork()
	if forked == nil {
		t.Fatalf("expected forked connection")
	}

	fc, ok := forked.(*Connect)
	if !ok {
		t.Fatalf("expected *Connect, got %T", forked)
	}

	if fc.Identifier != "missing" {
		t.Fatalf("expected identifier missing, got %q", fc.Identifier)
	}
}
