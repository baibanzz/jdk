package core

import (
	"github.com/baibanzz/jdk/core/internal/cron"
)

type Cron = cron.Cron

type CronOption = cron.Option

var (
	WithSeconds          = cron.WithSeconds
	WithLocation         = cron.WithLocation
	WithChain            = cron.WithChain
	WithLogger           = cron.WithLogger
	SkipIfStillRunning   = cron.SkipIfStillRunning
	Recover              = cron.Recover
)

func NewCron(opts ...CronOption) *Cron {
	return cron.New(opts...)
}
