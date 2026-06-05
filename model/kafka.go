package model

// Kafka 配置
//
// yaml 示例:
//
//	addrs:
//	  - kafka:29092
//	topic: my-topic
//	group: my-group
//	username: ""
//	password: ""
//	tls:
//	  enable: false
//	  caFile: ""
//	  certFile: ""
//	  keyFile: ""
//	  skipVerify: false
type Kafka struct {
	// Addrs Kafka 地址列表
	// yaml: addrs:
	//         - kafka:29092
	Addrs []string `json:"addrs" yaml:"addrs"`

	// Topic 默认主题
	// yaml: topic: my-topic
	Topic string `json:"topic" yaml:"topic"`

	// Group 消费者组
	// yaml: group: my-group
	Group string `json:"group" yaml:"group"`

	// Username SASL 用户名（可选）
	// yaml: username: ""
	Username string `json:"username" yaml:"username"`

	// Password SASL 密码（可选）
	// yaml: password: ""
	Password string `json:"password" yaml:"password"`

	// TLS TLS 配置（可选）
	// yaml: tls:
	//         enable: false
	//         caFile: ""
	//         certFile: ""
	//         keyFile: ""
	//         skipVerify: false
	TLS TLS `json:"tls" yaml:"tls"`
}

// TLS TLS 配置
type TLS struct {
	Enable     bool   `json:"enable" yaml:"enable"`
	CAFile     string `json:"caFile" yaml:"caFile"`
	CertFile   string `json:"certFile" yaml:"certFile"`
	KeyFile    string `json:"keyFile" yaml:"keyFile"`
	SkipVerify bool   `json:"skipVerify" yaml:"skipVerify"`
}
