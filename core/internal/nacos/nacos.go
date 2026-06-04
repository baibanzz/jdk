package nacos

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
)

type Nacos struct {
	configClient config_client.IConfigClient
	namingClient naming_client.INamingClient
	cfg          vo.NacosClientParam
	mu           sync.RWMutex
	closed       bool
}

func New(host []string, port uint64, namespace, logDir, cacheDir, username, password string) (*Nacos, error) {
	// 构建服务端配置
	serverConfigs := make([]constant.ServerConfig, 0, len(host))
	for _, ip := range host {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: ip,
			Port:   port,
		})
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:  namespace,
		LogDir:       logDir,
		CacheDir:     cacheDir,
		Username:     username,
		Password:     password,
		TimeoutMs:    10 * 1000,
		BeatInterval: 5 * 1000,
	}

	param := vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(param)
	if err != nil {
		return nil, fmt.Errorf("创建 Nacos 配置客户端失败: %w", err)
	}

	// 创建服务发现客户端
	namingClient, err := clients.NewNamingClient(param)
	if err != nil {
		configClient.CloseClient()
		return nil, fmt.Errorf("创建 Nacos 服务发现客户端失败: %w", err)
	}

	return &Nacos{
		configClient: configClient,
		namingClient: namingClient,
		cfg:          param,
	}, nil
}

// GetConfig 获取配置，返回原始 bytes
func (n *Nacos) GetConfig(dataId, group string) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil, fmt.Errorf("Nacos 客户端已关闭")
	}

	content, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Nacos 配置失败(dataId=%s, group=%s): %w", dataId, group, err)
	}

	return []byte(content), nil
}

// GetConfigToStruct 获取配置并反序列化到目标结构体
func (n *Nacos) GetConfigToStruct(dataId, group string, target any) error {
	data, err := n.GetConfig(dataId, group)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}
	if err = yaml.Unmarshal(data, target); err != nil {
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("Nacos 配置反序列化失败(dataId=%s, group=%s): %w", dataId, group, err)
		}
	}

	return nil
}

// ListenConfig 监听配置变更
// onChange 回调参数为变更后的配置内容(string)
func (n *Nacos) ListenConfig(dataId, group string, onChange func(string)) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return fmt.Errorf("Nacos 客户端已关闭")
	}

	return n.configClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			if onChange != nil {
				onChange(data)
			}
		},
	})
}

// CancelListenConfig 取消监听配置变更
func (n *Nacos) CancelListenConfig(dataId, group string) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return fmt.Errorf("Nacos 客户端已关闭")
	}

	return n.configClient.CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
}

// PushConfig 发布配置
func (n *Nacos) PushConfig(dataId, group string, content any) (bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	out, err := yaml.Marshal(content)
	if err != nil {
		return false, err
	}
	if n.closed {
		return false, fmt.Errorf("Nacos 客户端已关闭")
	}

	return n.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: string(out),
		Type:    "yaml",
	})
}

// DeleteConfig 删除配置
func (n *Nacos) DeleteConfig(dataId, group string) (bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return false, fmt.Errorf("Nacos 客户端已关闭")
	}

	return n.configClient.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
}

// RegisterService 注册服务实例
func (n *Nacos) RegisterService(serviceName, ip string, port int, opts ...RegisterOption) (bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return false, fmt.Errorf("Nacos 客户端已关闭")
	}

	param := vo.RegisterInstanceParam{
		ServiceName: serviceName,
		Ip:          ip,
		Port:        uint64(port),
		Weight:      1,
		Enable:      true,
		Healthy:     true,
	}

	for _, opt := range opts {
		opt(&param)
	}

	return n.namingClient.RegisterInstance(param)
}

// DeregisterService 注销服务实例
func (n *Nacos) DeregisterService(serviceName, ip string, port int) (bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return false, fmt.Errorf("Nacos 客户端已关闭")
	}

	return n.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: serviceName,
		Ip:          ip,
		Port:        uint64(port),
	})
}

// SelectInstances 获取服务实例列表（只返回健康的实例）
func (n *Nacos) SelectInstances(serviceName string, opts ...SelectOption) ([]model.Instance, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil, fmt.Errorf("Nacos 客户端已关闭")
	}

	param := vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	}

	for _, opt := range opts {
		opt(&param)
	}

	return n.namingClient.SelectInstances(param)
}

// SelectAllInstances 获取所有服务实例（包含不健康的）
func (n *Nacos) SelectAllInstances(serviceName string, opts ...SelectAllOption) ([]model.Instance, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil, fmt.Errorf("Nacos 客户端已关闭")
	}

	param := vo.SelectAllInstancesParam{
		ServiceName: serviceName,
	}

	for _, opt := range opts {
		opt(&param)
	}

	return n.namingClient.SelectAllInstances(param)
}

// SelectOneHealthyInstance 获取一个健康的服务实例（WRR 负载均衡）
func (n *Nacos) SelectOneHealthyInstance(serviceName string, opts ...SelectOneOption) (*model.Instance, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil, fmt.Errorf("Nacos 客户端已关闭")
	}

	param := vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	}

	for _, opt := range opts {
		opt(&param)
	}

	return n.namingClient.SelectOneHealthyInstance(param)
}

// Subscribe 订阅服务变更事件
func (n *Nacos) Subscribe(serviceName string, callback func(services []model.Instance, err error), opts ...SubscribeOption) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return fmt.Errorf("Nacos 客户端已关闭")
	}

	param := &vo.SubscribeParam{
		ServiceName:       serviceName,
		SubscribeCallback: callback,
	}

	for _, opt := range opts {
		opt(param)
	}

	return n.namingClient.Subscribe(param)
}

// Unsubscribe 取消订阅服务变更事件
func (n *Nacos) Unsubscribe(serviceName string, callback func(services []model.Instance, err error), opts ...SubscribeOption) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return fmt.Errorf("Nacos 客户端已关闭")
	}

	param := &vo.SubscribeParam{
		ServiceName:       serviceName,
		SubscribeCallback: callback,
	}

	for _, opt := range opts {
		opt(param)
	}

	return n.namingClient.Unsubscribe(param)
}

// Close 关闭 Nacos 客户端
func (n *Nacos) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return
	}
	n.closed = true

	n.configClient.CloseClient()
	n.namingClient.CloseClient()
}

// ========== 可选参数函数 ==========

type RegisterOption func(*vo.RegisterInstanceParam)
type SelectOption func(*vo.SelectInstancesParam)
type SelectAllOption func(*vo.SelectAllInstancesParam)
type SelectOneOption func(*vo.SelectOneHealthInstanceParam)
type SubscribeOption func(*vo.SubscribeParam)

// WithGroup 设置分组
func WithGroup(group string) RegisterOption {
	return func(p *vo.RegisterInstanceParam) {
		p.GroupName = group
	}
}

// WithGroupSelect 设置分组（查询用）
func WithGroupSelect(group string) SelectOption {
	return func(p *vo.SelectInstancesParam) {
		p.GroupName = group
	}
}

// WithGroupSelectAll 设置分组（查询所有用）
func WithGroupSelectAll(group string) SelectAllOption {
	return func(p *vo.SelectAllInstancesParam) {
		p.GroupName = group
	}
}

// WithGroupSelectOne 设置分组（查询单个用）
func WithGroupSelectOne(group string) SelectOneOption {
	return func(p *vo.SelectOneHealthInstanceParam) {
		p.GroupName = group
	}
}

// WithGroupSubscribe 设置分组（订阅用）
func WithGroupSubscribe(group string) SubscribeOption {
	return func(p *vo.SubscribeParam) {
		p.GroupName = group
	}
}

// WithWeight 设置权重（仅用于注册）
func WithWeight(weight float64) RegisterOption {
	return func(p *vo.RegisterInstanceParam) {
		p.Weight = weight
	}
}

// WithCluster 设置集群名（注册用）
func WithCluster(cluster string) RegisterOption {
	return func(p *vo.RegisterInstanceParam) {
		p.ClusterName = cluster
	}
}

// WithClusterSelect 设置集群名（查询用）
func WithClusterSelect(cluster string) SelectOption {
	return func(p *vo.SelectInstancesParam) {
		p.Clusters = []string{cluster}
	}
}

// WithClusterSelectAll 设置集群名（查询所有用）
func WithClusterSelectAll(cluster string) SelectAllOption {
	return func(p *vo.SelectAllInstancesParam) {
		p.Clusters = []string{cluster}
	}
}

// WithClusterSelectOne 设置集群名（查询单个用）
func WithClusterSelectOne(cluster string) SelectOneOption {
	return func(p *vo.SelectOneHealthInstanceParam) {
		p.Clusters = []string{cluster}
	}
}

// WithClusterSubscribe 设置集群名（订阅用）
func WithClusterSubscribe(cluster string) SubscribeOption {
	return func(p *vo.SubscribeParam) {
		p.Clusters = []string{cluster}
	}
}

// WithMetadata 设置元数据（仅用于注册）
func WithMetadata(metadata map[string]string) RegisterOption {
	return func(p *vo.RegisterInstanceParam) {
		p.Metadata = metadata
	}
}

// WithHealthyOnly 设置是否只返回健康实例（仅用于查询）
func WithHealthyOnly(healthyOnly bool) SelectOption {
	return func(p *vo.SelectInstancesParam) {
		p.HealthyOnly = healthyOnly
	}
}
