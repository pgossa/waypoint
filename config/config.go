package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StorageType string

const (
	StorageJSON   StorageType = "json"
	StorageSQLite StorageType = "sqlite"
)

type Config struct {
	Storage StorageType `json:"storage"`
}

var loaded *Config

func Load() (*Config, error) {
	if loaded != nil {
		return loaded, nil
	}

	path, err := configPath()
	if err != nil {
		return nil, err
	}

	// default config if file doesn't exist yet
	if _, err := os.Stat(path); os.IsNotExist(err) {
		loaded = &Config{Storage: StorageJSON}
		return loaded, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	loaded = &cfg
	return loaded, nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "waypoint", "config.json"), nil
}

func Reset() {
	loaded = nil
}

func DataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "waypoint", "data.json"), nil
}

func BasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "waypoint"), nil
}

func GetLastPrint() (time.Time, error) {
	base, err := BasePath()
	if err != nil {
		return time.Time{}, err
	}
	path := filepath.Join(base, ".last_print")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
}

func SetLastPrint() error {
	base, err := BasePath()
	if err != nil {
		return err
	}
	path := filepath.Join(base, ".last_print")
	return os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)), 0o644)
}

func IsInitialized() bool {
	path, err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return !os.IsNotExist(err)
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	loaded = cfg
	return json.NewEncoder(f).Encode(cfg)
}
