package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Option cron 选项类型
type Option = cron.Option

// Cron 定时任务调度器
type Cron struct {
	*cron.Cron
}

// New 创建定时任务调度器
// opts: 可选参数，如 cron.WithSeconds() 支持秒级精度
func New(opts ...cron.Option) *Cron {
	return &Cron{
		Cron: cron.New(opts...),
	}
}

// AddFunc 添加 cron 表达式任务
// spec: cron 表达式（如 "*/5 * * * *" 每5分钟，或 "*/5 * * * * *" 每5秒）
// cmd: 要执行的函数
// 返回: 任务 ID
func (c *Cron) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return c.Cron.AddFunc(spec, cmd)
}

// AddJob 添加实现了 cron.Job 接口的任务
// spec: cron 表达式
// job: 实现了 Run() 方法的任务
// 返回: 任务 ID
func (c *Cron) AddJob(spec string, job cron.Job) (cron.EntryID, error) {
	return c.Cron.AddJob(spec, job)
}

// Remove 移除指定 ID 的任务
func (c *Cron) Remove(id cron.EntryID) {
	c.Cron.Remove(id)
}

// Start 启动调度器（在独立 goroutine 中运行）
func (c *Cron) Start() {
	c.Cron.Start()
}

// Stop 停止调度器（返回一个 channel，等待所有正在执行的任务完成）
func (c *Cron) Stop() {
	<-c.Cron.Stop().Done()
}

// Entries 获取所有任务列表
func (c *Cron) Entries() []cron.Entry {
	return c.Cron.Entries()
}

// Entry 获取指定 ID 的任务
func (c *Cron) Entry(id cron.EntryID) cron.Entry {
	return c.Cron.Entry(id)
}

// ========== 便捷 Option 函数 ==========

// WithSeconds 支持秒级精度（6 位 cron 表达式）
// 标准 cron 是 5 位（分 时 日 月 周）
// 启用后是 6 位（秒 分 时 日 月 周）
func WithSeconds() cron.Option {
	return cron.WithSeconds()
}

// WithLocation 设置时区
func WithLocation(loc *time.Location) cron.Option {
	return cron.WithLocation(loc)
}

// WithChain 设置任务链（如 panic 恢复、日志记录等）
func WithChain(chain ...cron.JobWrapper) cron.Option {
	return cron.WithChain(chain...)
}

// WithLogger 设置自定义日志记录器
func WithLogger(logger cron.Logger) cron.Option {
	return cron.WithLogger(logger)
}

// SkipIfStillRunning 如果上次任务还没执行完，跳过本次执行
func SkipIfStillRunning() cron.JobWrapper {
	return cron.SkipIfStillRunning(cron.DefaultLogger)
}

// Recover 捕获任务 panic 并记录日志
func Recover() cron.JobWrapper {
	return cron.Recover(cron.DefaultLogger)
}
