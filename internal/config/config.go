package config

import (
	"github.com/burenotti/redis_impl/pkg/conf"
	"time"
)

type Config struct {
	Server struct {
		Host            string
		Port            int
		ShutdownTimeout time.Duration
		MaxConnections  int
	}
}

func Load(filePath string) (cfg *Config, err error) {
	cfg = &Config{}
	err = conf.BindFile(cfg, filePath)
	return cfg, err
}

func MustLoad(filePath string) *Config {
	cfg, err := Load(filePath)
	if err != nil {
		panic(err)
	}
	return cfg
}
