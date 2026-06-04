package core

import (
	"github.com/baibanzz/jdk/core/internal/nacos"
)

type Nacos = nacos.Nacos

// NacosConfig Nacos 配置
//
// yaml 示例:
//
//	nacos:
//	  host:
//	    - 192.168.1.100
//	  port: 8848
//	  username: nacos
//	  password: nacos
//	  namespace: public
//	  logDir: /tmp/nacos/log
//	  cacheDir: /tmp/nacos/cache
type NacosConfig struct {
	// Nacos 服务地址列表（支持集群）
	// yaml: host:
	//         - 192.168.1.100
	//         - 192.168.1.101
	Host []string `json:"host" yaml:"host"`

	// Nacos 服务端口，默认 8848
	// yaml: port: 8848
	Port uint64 `json:"port" yaml:"port"`

	// Nacos 登录用户名
	// yaml: username: nacos
	Username string `json:"username" yaml:"username"`

	// Nacos 登录密码
	// yaml: password: nacos
	Password string `json:"password" yaml:"password"`

	// Nacos 命名空间 ID，public 命名空间填空字符串
	// yaml: namespace: ""
	NameSpace string `json:"namespace" yaml:"namespace"`

	// Nacos 客户端日志目录
	// yaml: logDir: /tmp/nacos/log
	LogDir string `json:"logDir" yaml:"logDir"`

	// Nacos 客户端缓存目录
	// yaml: cacheDir: /tmp/nacos/cache
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`
}

func NewNacos(config NacosConfig) (*Nacos, error) {
	if config.CacheDir == "" {
		config.CacheDir = "./tmp/nacos/cache"
	}
	if config.LogDir == "" {
		config.LogDir = "./tmp/nacos/log"
	}
	return nacos.New(config.Host, config.Port, config.NameSpace, config.LogDir, config.CacheDir, config.Username, config.Password)
}
