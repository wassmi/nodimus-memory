package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
[server]
port = 8080
bind = "127.0.0.1"
timeout = 30

[storage]
data_dir = "/tmp/nodimus"

[logger]
level = "debug"
file = "test.log"
max_size = 10
max_backups = 2
max_age = 7
compress = false
`
	tmpfile, err := ioutil.TempFile("", "config_test_*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected logger level debug, got %s", cfg.Logger.Level)
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Server.Port != 4000 {
		t.Errorf("Expected default server port 4000, got %d", cfg.Server.Port)
	}
	if cfg.Storage.DataDir != "~/.nodimus" {
		t.Errorf("Expected default data dir ~/.nodimus, got %s", cfg.Storage.DataDir)
	}
}

func TestExpandDataDir(t *testing.T) {
	cfg := Default()

	// Test with tilde expansion
	expandedPath, err := cfg.ExpandDataDir()
	if err != nil {
		t.Fatalf("ExpandDataDir failed: %v", err)
	}
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".nodimus")
	if expandedPath != expectedPath {
		t.Errorf("Expected expanded path %s, got %s", expectedPath, expandedPath)
	}

	// Test with absolute path
	cfg.Storage.DataDir = "/var/lib/nodimus"
	expandedPath, err = cfg.ExpandDataDir()
	if err != nil {
		t.Fatalf("ExpandDataDir failed: %v", err)
	}
	if expandedPath != "/var/lib/nodimus" {
		t.Errorf("Expected expanded path /var/lib/nodimus, got %s", expandedPath)
	}

	// Test with empty path
	cfg.Storage.DataDir = ""
	expandedPath, err = cfg.ExpandDataDir()
	if err != nil {
		t.Fatalf("ExpandDataDir failed: %v", err)
	}
	if expandedPath != "" {
		t.Errorf("Expected empty expanded path, got %s", expandedPath)
	}
}
