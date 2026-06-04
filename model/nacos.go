package model

// Nacos 配置
//
// yaml 示例:
//
//	host:
//	  - 192.168.1.100
//	port: 8848
//	grpcPort: 9848
//	username: nacos
//	password: nacos
//	namespace: public
//	logDir: /tmp/nacos/log
//	cacheDir: /tmp/nacos/cache
type Nacos struct {
	// Host 服务地址列表（支持集群）
	// yaml: host:
	//         - 192.168.1.100
	//         - 192.168.1.101
	Host []string `json:"host" yaml:"host"`

	// Port 服务端口，默认 8848
	// yaml: port: 8848
	Port uint64 `json:"port" yaml:"port"`

	// GrpcPort gRPC 端口，默认 port + 1000（即 9848）
	// yaml: grpcPort: 9848
	GrpcPort uint64 `json:"grpcPort" yaml:"grpcPort"`

	// Username 登录用户名
	// yaml: username: nacos
	Username string `json:"username" yaml:"username"`

	// Password 登录密码
	// yaml: password: nacos
	Password string `json:"password" yaml:"password"`

	// NameSpace 命名空间 ID，public 命名空间填空字符串
	// yaml: namespace: ""
	NameSpace string `json:"namespace" yaml:"namespace"`

	// LogDir 客户端日志目录
	// yaml: logDir: /tmp/nacos/log
	LogDir string `json:"logDir" yaml:"logDir"`

	// CacheDir 客户端缓存目录
	// yaml: cacheDir: /tmp/nacos/cache
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`
}
