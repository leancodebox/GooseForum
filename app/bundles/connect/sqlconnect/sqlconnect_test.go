package sqlconnect

import "testing"

func TestTestConfigUsesSingleConnectionInMemorySQLite(t *testing.T) {
	cfg := TestConfig()

	if cfg.Connection != "sqlite" {
		t.Fatalf("Connection = %q, want sqlite", cfg.Connection)
	}
	if cfg.DbPath != ":memory:" {
		t.Fatalf("DbPath = %q, want :memory:", cfg.DbPath)
	}
	if cfg.MaxOpenConnections != 1 {
		t.Fatalf("MaxOpenConnections = %d, want 1", cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections != 1 {
		t.Fatalf("MaxIdleConnections = %d, want 1", cfg.MaxIdleConnections)
	}
}
