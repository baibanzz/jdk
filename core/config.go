package core

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func NewConfig(file string, Config any) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(file)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(Config); err != nil {
		return nil, err
	}
	return v, nil
}

func NewConfigListen[T any](file string, Config T, f func(t T)) (*viper.Viper, error) {
	v, err := NewConfig(file, Config)
	if err != nil {
		return nil, err
	}
	v.OnConfigChange(func(in fsnotify.Event) {
		if err := v.Unmarshal(Config); err != nil {
			return
		}
		if f != nil {
			f(Config)
		}
	})

	v.WatchConfig()
	return v, nil
}
