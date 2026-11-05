package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-sql-driver/mysql"
)

// clearEnv unsets the DB_* variables for the duration of a test so that the
// developer's own environment cannot influence the result.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{"DB_USERNAME", "DB_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME", "LISTEN_ADDR"} {
		t.Setenv(name, "") // registers restoration of the original value
		os.Unsetenv(name)
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	clearEnv(t)
	path := filepath.Join(t.TempDir(), "db.cfg")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigFromFile(t *testing.T) {
	path := writeConfig(t, `{"DB_USERNAME":"u","DB_PASSWORD":"p","DB_HOST":"db","DB_PORT":"3307","DB_NAME":"n"}`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBUser != "u" || cfg.DBHost != "db" || cfg.DBPort != "3307" || cfg.DBName != "n" {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr default = %q, want :8080", cfg.ListenAddr)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	path := writeConfig(t, `{"DB_USERNAME":"u","DB_NAME":"n","DB_HOST":"file-host"}`)
	t.Setenv("DB_HOST", "env-host")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBHost != "env-host" {
		t.Errorf("DBHost = %q, want env-host", cfg.DBHost)
	}
}

func TestMissingFileUsesEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_USERNAME", "u")
	t.Setenv("DB_NAME", "n")
	if _, err := LoadConfig(filepath.Join(t.TempDir(), "absent.cfg")); err != nil {
		t.Fatalf("expected env-only config to load, got %v", err)
	}
}

func TestRequiredFields(t *testing.T) {
	path := writeConfig(t, `{"DB_PASSWORD":"p"}`)
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("expected an error when DB_USERNAME/DB_NAME are missing")
	}
}

func TestDSNRoundTrip(t *testing.T) {
	cfg := Config{DBUser: "u", DBPassword: "p@ss:w/rd", DBHost: "h", DBPort: "3306", DBName: "n"}
	parsed, err := mysql.ParseDSN(cfg.DSN())
	if err != nil {
		t.Fatalf("ParseDSN: %v", err)
	}
	if parsed.User != "u" || parsed.Passwd != "p@ss:w/rd" || parsed.Addr != "h:3306" || parsed.DBName != "n" {
		t.Errorf("DSN did not round-trip: %+v", parsed)
	}
	if !parsed.ParseTime {
		t.Errorf("parseTime should be enabled")
	}
}
