package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/adrg/xdg"
	"go.yaml.in/yaml/v4"
)

const (
	AppName = "kagi-agent"
)

var (
	AppConfigFile string
	AppCacheFile  string

	mainConfigFile = "config.yaml"
	cacheFile      = "cache.yaml"
)

// Sets AppConfigDir and AppConfigFile
func initXDGLocs() error {
	var err error

	AppConfigFile, err = xdg.ConfigFile(filepath.Join(AppName, mainConfigFile))
	if err != nil {
		return fmt.Errorf("xdg.ConfigFile(): %w", err)
	}

	AppCacheFile, err = xdg.CacheFile(filepath.Join(AppName, cacheFile))
	if err != nil {
		return fmt.Errorf("xdg.CacheFile(): %w", err)
	}

	return nil
}

type Config struct {
	// https://github.com/charmbracelet/log/blob/main/level.go
	LogLevelStr string    `yaml:"loglevel"`
	LogLevel    log.Level `yaml:"-"`

	SessionToken string

	AlwaysAskForModel           bool
	AlwaysAskForReasoningEffort bool
}

type Cache struct {
	LastModel           string
	LastReasoningEffort string
}

// Attempt to read from disk
func Load() (*Config, error) {
	if err := initXDGLocs(); err != nil {
		return nil, err
	}

	_, err := os.Stat(AppConfigFile)
	if err != nil {
		return nil, err
	}

	cfgBytes, err := os.ReadFile(AppConfigFile)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile(AppConfigFile): %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal(): %w", err)
	}

	// validate fields
	cfg.LogLevel, err = log.ParseLevel(cfg.LogLevelStr)
	if err != nil {
		return &cfg, fmt.Errorf("log.ParseLevel(cfg.LogLevelStr): %w", err)
	}

	return &cfg, nil
}

// New saves a new config
func New(sessionToken string) (*Config, error) {
	cfg := Config{
		LogLevelStr: log.ErrorLevel.String(),
		LogLevel:    log.ErrorLevel,

		SessionToken:                sessionToken,
		AlwaysAskForModel:           true,
		AlwaysAskForReasoningEffort: true,
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	yamlBytes, err := yaml.Marshal(*c)
	if err != nil {
		return fmt.Errorf("yaml.Marshal(): %w", err)
	}

	err = os.WriteFile(AppConfigFile, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving config to disk: %w", err)
	}

	return nil
}

func DeleteConfig() error {
	if err := os.Remove(AppConfigFile); err != nil {
		return fmt.Errorf("os.Remove(): %w", err)
	}
	return nil
}

func LoadCache() (*Cache, error) {
	cacheBytes, err := os.ReadFile(AppCacheFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Cache{}, nil
		}
		return nil, fmt.Errorf("os.ReadFile(AppCacheFile): %w", err)
	}

	var cache Cache
	err = yaml.Unmarshal(cacheBytes, &cache)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal(): %w", err)
	}

	return &cache, nil
}

func (c *Cache) Save() error {
	yamlBytes, err := yaml.Marshal(*c)
	if err != nil {
		return fmt.Errorf("yaml.Marshal(): %w", err)
	}

	err = os.WriteFile(AppCacheFile, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving cache to disk: %w", err)
	}

	return nil
}

func DeleteCache() error {
	if err := os.Remove(AppCacheFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("os.Remove(): %w", err)
	}
	return nil
}
