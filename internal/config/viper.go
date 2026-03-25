package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	EnvPrefix = "PEACOCK"
)

func NewViper(explicitConfigPath, userConfigDir string) (*viper.Viper, error) {
	v := viper.New()
	var defaults map[string]any
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "mapstructure", Result: &defaults})
	if err != nil {
		return nil, fmt.Errorf("mapstructure decoder: %w", err)
	}
	if err := dec.Decode(DefaultConfig()); err != nil {
		return nil, fmt.Errorf("encode default config: %w", err)
	}
	if err := v.MergeConfigMap(defaults); err != nil {
		return nil, fmt.Errorf("merge default config: %w", err)
	}
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if explicitConfigPath != "" {
		v.SetConfigFile(explicitConfigPath)
		return v, nil
	}

	configDir, err := DefaultConfigDir(userConfigDir)
	if err != nil {
		return nil, err
	}

	v.AddConfigPath(configDir)
	v.SetConfigName(DefaultConfigBasename)
	v.SetConfigType(DefaultConfigExtension)

	return v, nil
}

func Load(v *viper.Viper) (Config, error) {
	if err := v.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
