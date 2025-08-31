package config

import (
	"os"
	"path/filepath"
	"testing"

	"crypto/tls"

	yaml "gopkg.in/yaml.v3"
)

type MockEnvMapper struct {
	EnvVars map[string]string
}

func NewMockEnvMapper() *MockEnvMapper {
	return &MockEnvMapper{EnvVars: make(map[string]string)}
}

func (m *MockEnvMapper) GetEnv(key string) string {
	return m.EnvVars[key]
}

func (m *MockEnvMapper) SetEnv(key, value string) error {
	if value == "" {
		delete(m.EnvVars, key)
	} else {
		m.EnvVars[key] = value
	}
	return nil
}

func TestNewDefaultConfig(t *testing.T) {
	c := NewDefaultConfig()
	if c.Folder != defaultFolder {
		t.Fatalf("expected default folder %q got %q", defaultFolder, c.Folder)
	}
	if c.Port != defaultPort {
		t.Fatalf("expected default port %d got %d", defaultPort, c.Port)
	}
}

func TestLoadConfig_FromEnvOnly(t *testing.T) {
	tmpDir := t.TempDir()
	env := NewMockEnvMapper()
	env.SetEnv(folderKey, tmpDir)
	cfg, err := LoadConfig("", env)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Folder != tmpDir {
		t.Fatalf("expected folder %q got %q", tmpDir, cfg.Folder)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "cfg.yaml")

	want := &Config{
		Folder:        "/from/file",
		Host:          "example.test",
		Port:          4242,
		TLSMinVersStr: "tls12",
	}

	b, err := yaml.Marshal(want)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}
	if err := os.WriteFile(cfgFile, b, 0600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	cfg, err := LoadConfig(cfgFile, NewMockEnvMapper())
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Folder != want.Folder {
		t.Fatalf("expected folder %q got %q", want.Folder, cfg.Folder)
	}
	if cfg.Port != want.Port {
		t.Fatalf("expected port %d got %d", want.Port, cfg.Port)
	}
	if cfg.Host != want.Host {
		t.Fatalf("expected host %q got %q", want.Host, cfg.Host)
	}
}

func TestValidate_TLSMismatch(t *testing.T) {
	c := NewDefaultConfig()
	c.TLSCert = "nonexistent.cert"
	c.TLSKey = ""
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error when only TLS_CERT set")
	}
	c = NewDefaultConfig()
	c.TLSCert = ""
	c.TLSKey = "nonexistent.key"
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error when only TLS_KEY set")
	}
}

func TestStrAsBool_and_TlsMinVersAsUint16(t *testing.T) {
	// strAsBool cases
	if v, err := strAsBool("1"); err != nil || !v {
		t.Fatalf("expected true from '1'")
	}
	if v, err := strAsBool("0"); err != nil || v {
		t.Fatalf("expected false from '0'")
	}
	if _, err := strAsBool("maybe"); err == nil {
		t.Fatalf("expected error for unknown boolean value")
	}

	// tlsMinVersAsUint16 cases
	if v, err := tlsMinVersAsUint16("tls12"); err != nil || v != tls.VersionTLS12 {
		t.Fatalf("expected tls12")
	}
	if _, err := tlsMinVersAsUint16("tls14"); err == nil {
		t.Fatalf("expected error for unknown tls version")
	}
}
