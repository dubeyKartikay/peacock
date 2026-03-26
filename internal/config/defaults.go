package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName                = "peacock"
	DefaultConfigBasename  = "config"
	DefaultConfigExtension = "yaml"
	DefaultConfigFilename  = DefaultConfigBasename + "." + DefaultConfigExtension

	FlagConfig = "config"

	defaultMaxEntries                = 5000
	defaultFilterPrompt              = "/ "
	defaultFilterPlaceholder         = "filter logs"
	defaultFilterCharLimit           = 256
	defaultScannerInitialBufferBytes = 64 * 1024
	defaultScannerMaxBufferBytes     = 1024 * 1024
	defaultFileTailLines             = 10
	defaultFilePoll                  = true
	defaultFileReopen                = false
	defaultPanelBorder               = "8"
	defaultStatusFG                  = "7"
	defaultFilterFG                  = "15"
	defaultFilterBG                  = "8"
	defaultTimestampFG               = "7"
	defaultTimestampFaint            = true
	defaultMessageFG                 = "15"
	defaultCallerFG                  = "4"
	defaultCallerFaint               = false
	defaultContextFG                 = "5"
	defaultContextFaint              = false
	defaultRawFG                     = "7"
	defaultLevelError                = "9"
	defaultLevelWarn                 = "11"
	defaultLevelInfo                 = "10"
	defaultLevelDebug                = "14"
	defaultLevelOther                = "7"
	defaultLevelBold                 = true
)

func DefaultConfigDir(userConfigDir string) (string, error) {
	resolvedUserConfigDir := userConfigDir
	if resolvedUserConfigDir == "" {
		var err error
		resolvedUserConfigDir, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve user config dir: %w", err)
		}
	}

	return filepath.Join(resolvedUserConfigDir, AppName), nil
}

func DefaultConfigPath(userConfigDir string) (string, error) {
	configDir, err := DefaultConfigDir(userConfigDir)
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, DefaultConfigFilename), nil
}
