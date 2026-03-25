package config

import (
	"path/filepath"
	"testing"
)

const (
	testConfigFilename = "config.yml"
)

func TestLoadReturnsDefaultsWithoutConfigFile(t *testing.T) {
	v, err := NewViper("", t.TempDir())
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	defaults := DefaultConfig()
	if cfg != defaults {
		t.Fatalf("expected defaults %#v, got %#v", defaults, cfg)
	}
}

func TestLoadReadsDefaultUserConfigPath(t *testing.T) {
	userConfigDir := t.TempDir()
	configPath, err := DefaultConfigPath(userConfigDir)
	if err != nil {
		t.Fatalf("default config path: %v", err)
	}

	writeConfigFile(t, configPath, `
buffer:
  max_entries: 42
theme:
  status_fg: "201"
`)

	v, err := NewViper("", userConfigDir)
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Buffer.MaxEntries != 42 {
		t.Fatalf("expected max entries 42, got %d", cfg.Buffer.MaxEntries)
	}
	if cfg.Theme.StatusFG != "201" {
		t.Fatalf("expected status fg 201, got %q", cfg.Theme.StatusFG)
	}
}

func TestLoadReadsExplicitConfigPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), testConfigFilename)
	writeConfigFile(t, configPath, `
input:
  filter_prompt: "search> "
source:
  file_tail_lines: 25
  file_reopen: true
`)

	v, err := NewViper(configPath, "")
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Input.FilterPrompt != "search> " {
		t.Fatalf("expected filter prompt search> , got %q", cfg.Input.FilterPrompt)
	}
	if cfg.Source.FileTailLines != 25 {
		t.Fatalf("expected file_tail_lines 25, got %d", cfg.Source.FileTailLines)
	}
	if !cfg.Source.FileReopen {
		t.Fatal("expected file_reopen to be true")
	}
}

func TestLoadPrefersEnvironmentOverConfigFile(t *testing.T) {
	userConfigDir := t.TempDir()
	configPath, err := DefaultConfigPath(userConfigDir)
	if err != nil {
		t.Fatalf("default config path: %v", err)
	}

	writeConfigFile(t, configPath, `
buffer:
  max_entries: 12
source:
  file_tail_lines: 7
theme:
  level_error: "196"
`)

	v, err := NewViper("", userConfigDir)
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	t.Setenv("PEACOCK_BUFFER_MAX_ENTRIES", "99")
	t.Setenv("PEACOCK_SOURCE_FILE_TAIL_LINES", "3")
	t.Setenv("PEACOCK_THEME_LEVEL_ERROR", "203")

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Buffer.MaxEntries != 99 {
		t.Fatalf("expected max entries 99, got %d", cfg.Buffer.MaxEntries)
	}
	if cfg.Source.FileTailLines != 3 {
		t.Fatalf("expected file tail lines 3, got %d", cfg.Source.FileTailLines)
	}
	if cfg.Theme.LevelError != "203" {
		t.Fatalf("expected level error 203, got %q", cfg.Theme.LevelError)
	}
}

func TestLoadValidatesValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), testConfigFilename)
	writeConfigFile(t, configPath, `
input:
  scanner_initial_buffer_bytes: 2048
  scanner_max_buffer_bytes: 1024
`)

	v, err := NewViper(configPath, "")
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	if _, err := Load(v); err == nil {
		t.Fatal("expected invalid config to fail validation")
	}
}

func TestLoadReadsEnvironmentWithoutConfigFile(t *testing.T) {
	v, err := NewViper("", t.TempDir())
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}

	t.Setenv("PEACOCK_INPUT_FILTER_PROMPT", "search> ")
	t.Setenv("PEACOCK_THEME_STATUS_BG", "240")

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Input.FilterPrompt != "search> " {
		t.Fatalf("expected filter prompt search> , got %q", cfg.Input.FilterPrompt)
	}
	if cfg.Theme.StatusBG != "240" {
		t.Fatalf("expected status bg 240, got %q", cfg.Theme.StatusBG)
	}
}
