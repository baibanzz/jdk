package core

import "github.com/baibanzz/jdk/core/internal/dll"

type Dll = dll.Lib

// LoadDll 动态加载DLL 支持windows Linux mac
func LoadDll(file, symbol string) (*Dll, error) {
	return dll.Load(file, symbol)
}
