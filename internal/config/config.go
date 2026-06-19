package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultBaseURL = "https://api.formatex.io"
	configDirName  = "formatex"
	configFileName = "config.json"
)

type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate config directory: %w", err)
	}
	return filepath.Join(dir, configDirName, configFileName), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// APIKey returns the effective API key: flag > env > config file.
func ResolveAPIKey(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if v := os.Getenv("FORMATEX_API_KEY"); v != "" {
		return v, nil
	}
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg.APIKey == "" {
		return "", fmt.Errorf("no API key found — run `formatex login` or set FORMATEX_API_KEY")
	}
	return cfg.APIKey, nil
}

func ResolveBaseURL(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv("FORMATEX_BASE_URL"); v != "" {
		return v
	}
	cfg, _ := Load()
	if cfg != nil && cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return DefaultBaseURL
}
