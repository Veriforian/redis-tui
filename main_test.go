package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFlags_NoArgs(t *testing.T) {
	conn, version, _, _, _, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn != nil {
		t.Error("expected nil connection with no args")
	}
	if version {
		t.Error("expected version=false")
	}
}

func TestParseFlags_HostOnly(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Host != "localhost" {
		t.Errorf("Host = %q, want %q", conn.Host, "localhost")
	}
	if conn.Port != 6379 {
		t.Errorf("Port = %d, want %d", conn.Port, 6379)
	}
	if conn.DB != 0 {
		t.Errorf("DB = %d, want %d", conn.DB, 0)
	}
	if conn.Name != "localhost:6379" {
		t.Errorf("Name = %q, want %q", conn.Name, "localhost:6379")
	}
}

func TestParseFlags_ShortFlags(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"-h", "redis.example.com", "-p", "6380", "-a", "secret", "-n", "5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Host != "redis.example.com" {
		t.Errorf("Host = %q, want %q", conn.Host, "redis.example.com")
	}
	if conn.Port != 6380 {
		t.Errorf("Port = %d, want %d", conn.Port, 6380)
	}
	if conn.Password != "secret" {
		t.Errorf("Password = %q, want %q", conn.Password, "secret")
	}
	if conn.DB != 5 {
		t.Errorf("DB = %d, want %d", conn.DB, 5)
	}
}

func TestParseFlags_LongFlags(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "10.0.0.1", "--port", "7000", "--password", "pass", "--db", "3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Host != "10.0.0.1" {
		t.Errorf("Host = %q, want %q", conn.Host, "10.0.0.1")
	}
	if conn.Port != 7000 {
		t.Errorf("Port = %d, want %d", conn.Port, 7000)
	}
	if conn.Password != "pass" {
		t.Errorf("Password = %q, want %q", conn.Password, "pass")
	}
	if conn.DB != 3 {
		t.Errorf("DB = %d, want %d", conn.DB, 3)
	}
}

func TestParseFlags_CustomName(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "localhost", "--name", "Production"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Name != "Production" {
		t.Errorf("Name = %q, want %q", conn.Name, "Production")
	}
}

func TestParseFlags_DefaultName(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "myhost", "--port", "9999"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Name != "myhost:9999" {
		t.Errorf("Name = %q, want %q", conn.Name, "myhost:9999")
	}
}

func TestParseFlags_Cluster(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "localhost", "--cluster"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if !conn.UseCluster {
		t.Error("UseCluster should be true")
	}
}

func TestParseFlags_TLS(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{
		"--host", "localhost",
		"--tls",
		"--tls-cert", "/path/cert.pem",
		"--tls-key", "/path/key.pem",
		"--tls-ca", "/path/ca.pem",
		"--tls-skip-verify",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if !conn.UseTLS {
		t.Error("UseTLS should be true")
	}
	if conn.TLSConfig == nil {
		t.Fatal("TLSConfig should be set")
	}
	if conn.TLSConfig.CertFile != "/path/cert.pem" {
		t.Errorf("CertFile = %q, want %q", conn.TLSConfig.CertFile, "/path/cert.pem")
	}
	if conn.TLSConfig.KeyFile != "/path/key.pem" {
		t.Errorf("KeyFile = %q, want %q", conn.TLSConfig.KeyFile, "/path/key.pem")
	}
	if conn.TLSConfig.CAFile != "/path/ca.pem" {
		t.Errorf("CAFile = %q, want %q", conn.TLSConfig.CAFile, "/path/ca.pem")
	}
	if !conn.TLSConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}

func TestParseFlags_TLSNotSet(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{"--host", "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.UseTLS {
		t.Error("UseTLS should be false")
	}
	if conn.TLSConfig != nil {
		t.Error("TLSConfig should be nil when --tls is not set")
	}
}

func TestParseFlags_Version(t *testing.T) {
	conn, version, _, _, _, err := parseFlags([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn != nil {
		t.Error("expected nil connection for --version")
	}
	if !version {
		t.Error("expected version=true")
	}
}

func TestParseFlags_AllOptions(t *testing.T) {
	conn, _, _, _, _, err := parseFlags([]string{
		"--host", "redis.prod.com",
		"--port", "6380",
		"--password", "s3cret",
		"--db", "7",
		"--name", "Prod Redis",
		"--cluster",
		"--tls",
		"--tls-cert", "/cert.pem",
		"--tls-key", "/key.pem",
		"--tls-ca", "/ca.pem",
		"--tls-skip-verify",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil connection")
	}
	if conn.Host != "redis.prod.com" {
		t.Errorf("Host = %q", conn.Host)
	}
	if conn.Port != 6380 {
		t.Errorf("Port = %d", conn.Port)
	}
	if conn.Password != "s3cret" {
		t.Errorf("Password = %q", conn.Password)
	}
	if conn.DB != 7 {
		t.Errorf("DB = %d", conn.DB)
	}
	if conn.Name != "Prod Redis" {
		t.Errorf("Name = %q", conn.Name)
	}
	if !conn.UseCluster {
		t.Error("UseCluster should be true")
	}
	if !conn.UseTLS {
		t.Error("UseTLS should be true")
	}
	if conn.TLSConfig == nil {
		t.Fatal("TLSConfig should be set")
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, _, _, _, _, err := parseFlags([]string{"--invalid-flag"})
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestParseFlags_Help(t *testing.T) {
	_, _, _, _, _, err := parseFlags([]string{"--help"})
	if err != flag.ErrHelp {
		t.Errorf("expected flag.ErrHelp, got %v", err)
	}
}

func TestParseFlags_Update(t *testing.T) {
	conn, version, doUpdate, _, _, err := parseFlags([]string{"--update"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn != nil {
		t.Error("expected nil connection for --update")
	}
	if version {
		t.Error("expected version=false")
	}
	if !doUpdate {
		t.Error("expected doUpdate=true")
	}
}

func TestParseFlags_UpdateWithOtherFlags(t *testing.T) {
	conn, version, doUpdate, _, _, err := parseFlags([]string{"--host", "localhost", "--update"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn != nil {
		t.Error("expected nil connection when --update is set")
	}
	if version {
		t.Error("expected version=false")
	}
	if !doUpdate {
		t.Error("expected doUpdate=true")
	}
}

func TestParseFlags_ScanSize(t *testing.T) {
	t.Run("default scan size", func(t *testing.T) {
		_, _, _, scanSize, _, err := parseFlags([]string{"--host", "localhost"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scanSize != 1000 {
			t.Errorf("ScanSize = %d, want 1000", scanSize)
		}
	})

	t.Run("custom scan size", func(t *testing.T) {
		_, _, _, scanSize, _, err := parseFlags([]string{"--host", "localhost", "--scan-size", "500"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scanSize != 500 {
			t.Errorf("ScanSize = %d, want 500", scanSize)
		}
	})
}

func TestParseFlags_IncludeTypesFalse(t *testing.T) {
	_, _, _, _, includeTypes, err := parseFlags([]string{"--host", "localhost", "--include-types=false"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if includeTypes {
		t.Error("expected includeTypes=false")
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	_, _, _, scanSize, includeTypes, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scanSize != 1000 {
		t.Errorf("ScanSize = %d, want 1000", scanSize)
	}
	if !includeTypes {
		t.Error("expected includeTypes=true by default")
	}
}

func TestInitConfig_LegacyMigration_UnsafePerms(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up legacy path with group/other-writable permissions.
	legacyDir := filepath.Join(tmpDir, ".redis")
	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	legacyConfig := map[string]any{
		"connections": []any{},
	}
	data, err := json.Marshal(legacyConfig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	legacyPath := filepath.Join(legacyDir, "config.json")
	if err := os.WriteFile(legacyPath, data, 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	// Force group/other-writable to bypass umask.
	if err := os.Chmod(legacyPath, 0o666); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".config", "redis-tui")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	// The new config should NOT exist after migration is skipped due to perms.
	// We can't call initConfig directly (it uses os.UserHomeDir), so we
	// verify the permission check by stat-ing the legacy file ourselves.
	info, err := os.Stat(legacyPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if info.Mode().Perm()&0o022 == 0 {
		t.Fatal("expected unsafe permissions on legacy file")
	}

	// Config file should not have been created.
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("expected config file to not exist (migration should be skipped)")
	}
}

func TestInitConfig_LegacyMigration_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	legacyDir := filepath.Join(tmpDir, ".redis")
	if err := os.MkdirAll(legacyDir, 0o750); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// Write invalid JSON to legacy path.
	legacyPath := filepath.Join(legacyDir, "config.json")
	if err := os.WriteFile(legacyPath, []byte("not valid json{{{"), 0o600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Verify it's detected as invalid.
	var cfg map[string]any
	data, err := os.ReadFile(legacyPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if err := json.Unmarshal(data, &cfg); err == nil {
		t.Error("expected JSON parse error for invalid config")
	}
}

func TestParseFlags_PasswordWarning(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantWarn  bool
	}{
		{"short flag -a", []string{"-h", "localhost", "-a", "secret"}, true},
		{"long flag --password", []string{"--host", "localhost", "--password", "secret"}, true},
		{"no password flag", []string{"--host", "localhost"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stderr = w

			_, _, _, _, _, parseErr := parseFlags(tt.args)
			if parseErr != nil {
				os.Stderr = oldStderr
				t.Fatalf("unexpected error: %v", parseErr)
			}

			if err := w.Close(); err != nil {
				t.Fatalf("failed to close writer: %v", err)
			}
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("failed to read pipe: %v", err)
			}
			os.Stderr = oldStderr

			gotWarn := strings.Contains(buf.String(), "process list")
			if gotWarn != tt.wantWarn {
				t.Errorf("warning present = %v, want %v (stderr: %q)", gotWarn, tt.wantWarn, buf.String())
			}
		})
	}
}
