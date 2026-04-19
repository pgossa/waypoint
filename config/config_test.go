package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"waypoint/config"
)

func writeConfig(t *testing.T, cfg config.Config) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(cfg); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return dir
}

func TestLoadDefaults(t *testing.T) {
	config.Reset()
	// point HOME to an empty temp dir so no config file exists
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Storage != config.StorageJSON {
		t.Errorf("got storage %q, want %q", cfg.Storage, config.StorageJSON)
	}
}

func TestLoadFromFile(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// create the expected config path
	configDir := filepath.Join(dir, ".config", "waypoint")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	path := filepath.Join(configDir, "config.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(config.Config{Storage: config.StorageSQLite}); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Storage != config.StorageSQLite {
		t.Errorf("got storage %q, want %q", cfg.Storage, config.StorageSQLite)
	}
}

func TestLoadCached(t *testing.T) {
	config.Reset()
	t.Setenv("HOME", t.TempDir())

	first, _ := config.Load()
	second, _ := config.Load()

	if first != second {
		t.Error("expected Load() to return the same cached instance")
	}
}

func TestDataPath(t *testing.T) {
	t.Setenv("HOME", "/tmp/test")
	path, err := config.DataPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "/tmp/test/.config/waypoint/data.json"
	if path != expected {
		t.Errorf("got path %q, want %q", path, expected)
	}
}
