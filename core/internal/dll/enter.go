package dll

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/ebitengine/purego"
)

// suffixed 匹配已带库后缀的路径,支持 .so 带版本号如 libfoo.so.1 / .so.1.2
var suffixed = regexp.MustCompile(`\.(dll|dylib|so)(\.\d+)*$`)

// Lib 动态库句柄
type Lib struct {
	handle uintptr
	symbol string
}

// resolve 按当前平台给库名补后缀: windows->.dll, linux->.so, darwin->.dylib
// 如果库名已带后缀(含 .so 带版本号,如 libfoo.so.1),则不重复拼接
func resolve(name string) string {
	if suffixed.MatchString(name) {
		return name
	}
	switch runtime.GOOS {
	case "windows":
		return name + ".dll"
	case "darwin":
		return name + ".dylib"
	default:
		return name + ".so"
	}
}

// Load 导入动态库(文件名,函数名),自动补平台后缀
// file: 库名,可带也可不带后缀(如 "sqlite3" 或 "sqlite3.so")
// symbol: 函数名
func Load(file, symbol string) (*Lib, error) {
	path := resolve(file)
	handle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, fmt.Errorf("打开动态库 %s 失败: %w", path, err)
	}
	return &Lib{handle: handle, symbol: symbol}, nil
}

// Bind 把符号绑定成 Go 可调用函数(nativeFn 必须传指针,如 &fn)
// 调用后 fn 就直接当普通 Go 函数调
func (l *Lib) Bind(nativeFn any) error {
	purego.RegisterLibFunc(nativeFn, l.handle, l.symbol)
	return nil
}

// Unload 卸载动态库
func (l *Lib) Unload() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	return purego.Dlclose(l.handle)
}
