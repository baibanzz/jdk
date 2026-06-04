package core

import (
	"github.com/baibanzz/jdk/core/internal/nacos"
	"github.com/baibanzz/jdk/model"
)

type Nacos = nacos.Nacos

func NewNacos(config model.Nacos) (*Nacos, error) {
	if config.CacheDir == "" {
		config.CacheDir = "./tmp/nacos/cache"
	}
	if config.LogDir == "" {
		config.LogDir = "./tmp/nacos/log"
	}
	return nacos.New(config.Host, config.Port, config.GrpcPort, config.NameSpace, config.LogDir, config.CacheDir, config.Username, config.Password)
}
