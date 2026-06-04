package core

import "github.com/baibanzz/jdk/core/internal/logs"

type (
	Logger     = logs.Logger
	Option     = logs.Option
	GormLogger = logs.GormLogger
)

var DefaultOption = logs.DefaultOption

func NewLog(opts ...Option) *Logger {
	var cfg Option
	if len(opts) == 0 {
		cfg = DefaultOption
	} else {
		cfg = opts[0]
	}
	return logs.NewLog(cfg)
}
