package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

var (
	ErrConfigNotLoaded = errors.New("config not loaded")
)

type Config struct {
	Server struct {
		Host            string        `yaml:"host" env-default:"localhost"`
		Port            int           `yaml:"port" env-default:"6379"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"5s"`
		MaxConnections  int           `yaml:"max_connections" env-default:"1024"`
	} `yaml:"server"`
}

func Load(filePath string) (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(filePath, cfg); err != nil {
		return nil, configNotLoadedErr("config not loaded: %w", err)
	}

	return cfg, nil
}

func MustLoad(filePath string) *Config {
	cfg, err := Load(filePath)
	if err != nil {
		panic(err)
	}
	return cfg
}

func configNotLoadedErr(format string, args ...any) error {
	return errors.Join(fmt.Errorf(format, args...), ErrConfigNotLoaded)
}
