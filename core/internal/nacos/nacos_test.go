package nacos

import (
	"log"
	"os"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"gopkg.in/yaml.v3"
)

var (
	nacosHost     = envOrDefault("NACOS_HOST", "127.0.0.1")
	nacosPort     = uint64(8848)
	nacosGrpcPort = uint64(9848)
	nacosUsername = envOrDefault("NACOS_USERNAME", "nacos")
	nacosPassword = envOrDefault("NACOS_PASSWORD", "nacos")
	nacosSpace    = envOrDefault("NACOS_NAMESPACE", "cex-qufc")
	testDataId    = envOrDefault("NACOS_TEST_DATA_ID", "demo.yaml")
	testGroup     = envOrDefault("NACOS_TEST_GROUP", "DEFAULT_GROUP")
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestNewNacos(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()
	t.Log("Nacos 客户端创建成功")
}

func TestNacos_GetConfig(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	data, err := nc.GetConfig(testDataId, testGroup)
	if err != nil {
		t.Fatalf("获取配置失败: %v", err)
	}
	t.Logf("获取配置成功, 内容长度: %d", len(data))
	if len(data) > 0 {
		t.Logf("配置内容: %s", string(data))
	}
}

func TestNacos_GetConfigToStruct(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	var result map[string]any
	err = nc.GetConfigToStruct(testDataId, testGroup, &result)
	if err != nil {
		t.Fatalf("获取配置并反序列化失败: %v", err)
	}
	t.Logf("反序列化结果: %+v", result)
}

func TestNacos_PublishAndDeleteConfig(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	var content = struct {
		Key string
		Val int
	}{
		"test-value",
		123,
	}
	// 发布配置
	ok, err := nc.PushConfig(testDataId, testGroup, content)
	if err != nil {
		t.Fatalf("发布配置失败: %v", err)
	}
	if !ok {
		t.Fatal("发布配置返回 false")
	}
	t.Log("发布配置成功")

	// 验证配置
	data, err := nc.GetConfig(testDataId, testGroup)
	if err != nil {
		t.Fatalf("获取配置失败: %v", err)
	}
	out, _ := yaml.Marshal(content)

	t.Logf("配置内容, 期望: %s, 实际: %s\n", string(out), string(data))

	// 删除配置
	ok, err = nc.DeleteConfig(testDataId, testGroup)
	if err != nil {
		t.Fatalf("删除配置失败: %v", err)
	}
	if !ok {
		t.Fatal("删除配置返回 false")
	}
	t.Log("删除配置成功")
}

func TestNacos_ListenConfig(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()
	nc.PushConfig(testDataId, testGroup, map[string]interface{}{
		"testKey": "testValue",
	})
	changed := make(chan string)

	err = nc.ListenConfig(testDataId, testGroup, func(data string) {
		changed <- data
	})
	if err != nil {
		t.Fatalf("监听配置失败: %v", err)
	}
	t.Log("监听配置成功，请手动在 Nacos 控制台修改配置以触发变更")
	log.Printf("配置变更: %s", <-changed)
}

func TestNacos_RegisterAndDeregisterService(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	serviceName := "test-go-service"
	ip := "127.0.0.1"
	port := 9848

	// 注册服务
	ok, err := nc.RegisterService(serviceName, ip, port)
	if err != nil {
		t.Fatalf("注册服务失败: %v", err)
	}
	if !ok {
		t.Fatal("注册服务返回 false")
	}
	t.Logf("注册服务成功: %s", serviceName)

	// 查询服务实例
	instances, err := nc.SelectInstances(serviceName)
	if err != nil {
		t.Fatalf("查询服务实例失败: %v", err)
	}
	t.Logf("查询到 %d 个服务实例", len(instances))
	for _, inst := range instances {
		t.Logf("  实例: %s:%d, 健康=%v, 启用=%v", inst.Ip, inst.Port, inst.Healthy, inst.Enable)
	}

	// 注销服务
	ok, err = nc.DeregisterService(serviceName, ip, port)
	if err != nil {
		t.Fatalf("注销服务失败: %v", err)
	}
	if !ok {
		t.Fatal("注销服务返回 false")
	}
	t.Logf("注销服务成功: %s", serviceName)
}

func TestNacos_SelectOneHealthyInstance(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	// 先注册一个服务
	serviceName := "test-healthy-service"
	_, _ = nc.RegisterService(serviceName, "127.0.0.1", 8081)

	// 获取一个健康实例
	inst, err := nc.SelectOneHealthyInstance(serviceName)
	if err != nil {
		t.Logf("获取健康实例失败(可能无可用实例): %v", err)
	} else if inst != nil {
		t.Logf("获取到健康实例: %s:%d", inst.Ip, inst.Port)
	}

	// 清理
	_, _ = nc.DeregisterService(serviceName, "127.0.0.1", 8081)
}

func TestNacos_Subscribe(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}
	defer nc.Close()

	serviceName := "test-subscribe-service"

	// 订阅服务变更
	err = nc.Subscribe(serviceName, func(services []model.Instance, err error) {
		if err != nil {
			t.Logf("订阅回调错误: %v", err)
			return
		}
		t.Logf("服务变更通知, 当前实例数: %d", len(services))
	})
	if err != nil {
		t.Fatalf("订阅服务失败: %v", err)
	}
	t.Log("订阅服务成功")

	// 取消订阅
	err = nc.Unsubscribe(serviceName, func(services []model.Instance, err error) {
	})
	if err != nil {
		t.Fatalf("取消订阅失败: %v", err)
	}
	t.Log("取消订阅成功")
}

func TestNacos_Close(t *testing.T) {
	nc, err := New(
		[]string{nacosHost},
		nacosPort,
		nacosGrpcPort,
		nacosSpace,
		"./tmp/nacos/log",
		"./tmp/nacos/cache",
		nacosUsername,
		nacosPassword,
	)
	if err != nil {
		t.Fatalf("创建 Nacos 客户端失败: %v", err)
	}

	nc.Close()
	t.Log("Nacos 客户端关闭成功")

	// 关闭后再操作应返回错误
	_, err = nc.GetConfig(testDataId, testGroup)
	if err == nil {
		t.Fatal("关闭后获取配置应返回错误")
	}
	t.Logf("关闭后操作返回预期错误: %v", err)
}
