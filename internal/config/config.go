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

func Load(filePath string) (res *Config, err error) {
	cfg, err := conf.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	res = &Config{}
	res.Server.Host = cfg.Get("bind").MustString("127.0.0.1")
	res.Server.Port = cfg.Get("port").MustInt(6379)
	res.Server.MaxConnections = cfg.Get("max_connection").MustInt(16)
	res.Server.ShutdownTimeout = time.Duration(cfg.Get("shutdown_timeout").MustInt(5)) * time.Second

	return
}

func MustLoad(filePath string) *Config {
	cfg, err := Load(filePath)
	if err != nil {
		panic(err)
	}
	return cfg
}
