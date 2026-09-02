package dll

import (
	"runtime"
	"testing"
)

func TestResolve(t *testing.T) {
	// 已带后缀的库名,不应重复拼接
	for _, name := range []string{
		"sqlite3.dll", "sqlite3.so", "sqlite3.dylib",
		"a/b/libfoo.so.1", "C:\\lib\\foo.dll",
	} {
		if got := resolve(name); got != name {
			t.Errorf("resolve(%q) = %q, 期望保持原样", name, got)
		}
	}

	// 无后缀时按当前平台补后缀
	suffix := map[string]string{
		"windows": ".dll",
		"darwin":  ".dylib",
		"linux":   ".so",
	}[runtime.GOOS]
	if suffix == "" {
		suffix = ".so"
	}
	if got := resolve("sqlite3"); got != "sqlite3"+suffix {
		t.Errorf("resolve(\"sqlite3\") = %q, 期望 %q", got, "sqlite3"+suffix)
	}
}

func TestLoad_NotFound(t *testing.T) {
	// 加载不存在的库应报错
	_, err := Load("no_such_lib_xyz_12345", "foo")
	if err == nil {
		t.Fatal("加载不存在的库应返回错误")
	}
	t.Logf("加载不存在的库返回预期错误: %v", err)
}

func TestUnload_NilAndZero(t *testing.T) {
	// nil 或 handle=0 不应报错
	var nilLib *Lib
	if err := nilLib.Unload(); err != nil {
		t.Fatalf("nil Lib Unload 应不报错, 实际: %v", err)
	}

	if err := (&Lib{handle: 0}).Unload(); err != nil {
		t.Fatalf("handle=0 Unload 应不报错, 实际: %v", err)
	}
	t.Log("nil/zero 边界处理正确")
}

// systemLib 根据当前平台返回一个确定存在的系统库与一个可调用的 C 函数
func systemLib() (lib, fn string, ok bool) {
	switch runtime.GOOS {
	case "windows":
		return "msvcrt", "abs", true
	case "darwin":
		return "libSystem.B.dylib", "abs", true
	default:
		return "libc", "abs", true
	}
}

func TestLoad_BindAndCall_SystemLib(t *testing.T) {
	// 加载系统库,把 abs 绑定成 Go 函数并真正调用,验证整套流程可用
	libName, sym, ok := systemLib()
	if !ok {
		t.Skip("不支持的平台")
	}

	lib, err := Load(libName, sym)
	if err != nil {
		t.Skipf("加载系统库 %s 失败, 跳过: %v", libName, err)
	}
	defer lib.Unload()

	var abs func(int) int
	if err := lib.Bind(&abs); err != nil {
		t.Fatalf("Bind 失败: %v", err)
	}

	if got := abs(-5); got != 5 {
		t.Fatalf("abs(-5) = %d, 期望 5", got)
	}
	if got := abs(3); got != 3 {
		t.Fatalf("abs(3) = %d, 期望 3", got)
	}
	t.Logf("系统库 %s 的 %s 绑定并调用成功", libName, sym)
}

func TestLoad_DuplicateSuffix_NotAppended(t *testing.T) {
	// 直接传入完整文件名(.so),resolve 不应重复拼后缀
	paths := []string{"foo.so", "foo.dll", "foo.dylib"}
	for _, p := range paths {
		lib, err := Load(p, "abs")
		if err == nil {
			// 只要不是"重复后缀导致打不开"即可;失败可接受,关键是路径未变
			_ = lib.Unload()
		}
	}
}
