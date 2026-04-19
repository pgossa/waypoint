package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"waypoint/config"
)

// =============================================================================
// BasePath
// =============================================================================

func TestBasePath(t *testing.T) {
	t.Setenv("HOME", "/tmp/test-home")
	path, err := config.BasePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "/tmp/test-home/.config/waypoint"
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

// =============================================================================
// IsInitialized
// =============================================================================

func TestIsInitialized_False_WhenNoConfigFile(t *testing.T) {
	config.Reset()
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(config.Reset)

	if config.IsInitialized() {
		t.Error("expected IsInitialized=false when config file does not exist")
	}
}

func TestIsInitialized_True_WhenConfigFileExists(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	// create the config file
	cfgDir := filepath.Join(dir, ".config", "waypoint")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"storage":"json"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if !config.IsInitialized() {
		t.Error("expected IsInitialized=true when config file exists")
	}
}

// =============================================================================
// Save
// =============================================================================

func TestSave_WritesConfigFile(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	cfg := &config.Config{Storage: config.StorageJSON}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// config file must now exist
	if !config.IsInitialized() {
		t.Error("expected config file to exist after Save")
	}
}

func TestSave_LoadReturnsTheSavedValue(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	if err := config.Save(&config.Config{Storage: config.StorageSQLite}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	config.Reset() // clear cache so Load reads from disk
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if loaded.Storage != config.StorageSQLite {
		t.Errorf("got storage %q, want %q", loaded.Storage, config.StorageSQLite)
	}
}

func TestSave_CreatesDirectoryIfMissing(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	// directory does NOT exist yet
	cfg := &config.Config{Storage: config.StorageJSON}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save with missing directory: %v", err)
	}
	if !config.IsInitialized() {
		t.Error("expected config to be initialized after Save")
	}
}

// =============================================================================
// GetLastPrint / SetLastPrint
// =============================================================================

func TestGetLastPrint_ReturnsZeroWhenFileAbsent(t *testing.T) {
	config.Reset()
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(config.Reset)

	last, err := config.GetLastPrint()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !last.IsZero() {
		t.Errorf("expected zero time when .last_print is absent, got %v", last)
	}
}

func TestSetLastPrint_WritesTimestamp(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	cfgDir := filepath.Join(dir, ".config", "waypoint")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	before := time.Now().Truncate(time.Second)

	if err := config.SetLastPrint(); err != nil {
		t.Fatalf("SetLastPrint: %v", err)
	}

	last, err := config.GetLastPrint()
	if err != nil {
		t.Fatalf("GetLastPrint: %v", err)
	}

	after := time.Now().Add(time.Second)
	if last.Before(before) || last.After(after) {
		t.Errorf("timestamp %v is outside expected window [%v, %v]", last, before, after)
	}
}

func TestSetLastPrint_OverwritesPreviousTimestamp(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	cfgDir := filepath.Join(dir, ".config", "waypoint")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// write once, then wait a bit, then write again
	if err := config.SetLastPrint(); err != nil {
		t.Fatal(err)
	}
	first, _ := config.GetLastPrint()

	time.Sleep(2 * time.Second)

	if err := config.SetLastPrint(); err != nil {
		t.Fatal(err)
	}
	second, _ := config.GetLastPrint()

	if !second.After(first) {
		t.Errorf("second timestamp (%v) should be after first (%v)", second, first)
	}
}

func TestSetLastPrint_RequiresBaseDirectory(t *testing.T) {
	config.Reset()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Cleanup(config.Reset)

	// SetLastPrint does NOT create the parent directory — it errors if missing.
	err := config.SetLastPrint()
	if err == nil {
		t.Error("expected error when base directory does not exist")
	}
}
