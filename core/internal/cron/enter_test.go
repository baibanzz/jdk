package cron

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCron(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("创建 Cron 失败")
	}
	t.Log("Cron 创建成功")
}

func TestCron_AddFunc(t *testing.T) {
	c := New(WithSeconds())
	c.Start()
	defer c.Stop()

	var count atomic.Int32

	id, err := c.AddFunc("* * * * * *", func() {
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("AddFunc 失败: %v", err)
	}
	t.Logf("添加任务成功, id=%d", id)

	// 等待 1.5 秒，应该至少执行 1 次
	time.Sleep(1500 * time.Millisecond)

	if count.Load() == 0 {
		t.Fatal("任务未执行")
	}
	t.Logf("任务执行了 %d 次", count.Load())

	// 移除任务
	c.Remove(id)
	prevCount := count.Load()

	time.Sleep(1100 * time.Millisecond)

	if count.Load() != prevCount {
		t.Fatal("移除任务后不应再执行")
	}
	t.Log("移除任务后不再执行")
}

type testJob struct {
	count *atomic.Int32
}

func (j *testJob) Run() {
	j.count.Add(1)
}

func TestCron_AddJob(t *testing.T) {
	c := New(WithSeconds())
	c.Start()
	defer c.Stop()

	var count atomic.Int32

	job := &testJob{count: &count}

	id, err := c.AddJob("* * * * * *", job)
	if err != nil {
		t.Fatalf("AddJob 失败: %v", err)
	}
	t.Logf("添加 Job 成功, id=%d", id)

	time.Sleep(5000 * time.Millisecond)

	if count.Load() == 0 {
		t.Fatal("Job 未执行")
	}
	t.Logf("Job 执行了 %d 次", count.Load())

	c.Remove(id)
}

func TestCron_Entries(t *testing.T) {
	c := New(WithSeconds())
	defer c.Stop()

	id1, _ := c.AddFunc("* * * * * *", func() {})
	id2, _ := c.AddFunc("* * * * * *", func() {})

	entries := c.Entries()
	if len(entries) != 2 {
		t.Fatalf("期望 2 个任务, 实际 %d", len(entries))
	}

	entry1 := c.Entry(id1)
	if entry1.ID != id1 {
		t.Fatalf("Entry ID 不匹配")
	}

	entry2 := c.Entry(id2)
	if entry2.ID != id2 {
		t.Fatalf("Entry ID 不匹配")
	}

	t.Logf("任务列表: %d 个", len(entries))
}

func TestCron_WithSeconds(t *testing.T) {
	c := New(WithSeconds())
	if c == nil {
		t.Fatal("创建带秒级精度的 Cron 失败")
	}
	t.Log("带秒级精度的 Cron 创建成功")
}

func TestCron_StartStop(t *testing.T) {
	c := New(WithSeconds())

	var count atomic.Int32
	_, _ = c.AddFunc("* * * * * *", func() {
		count.Add(1)
	})

	// 启动前不应执行
	time.Sleep(1100 * time.Millisecond)
	if count.Load() > 0 {
		t.Fatal("启动前不应执行任务")
	}

	// 启动
	c.Start()
	time.Sleep(1500 * time.Millisecond)

	if count.Load() == 0 {
		t.Fatal("启动后任务应执行")
	}
	t.Logf("启动后任务执行了 %d 次", count.Load())

	// 停止
	c.Stop()
	prevCount := count.Load()

	time.Sleep(1100 * time.Millisecond)

	if count.Load() != prevCount {
		t.Fatal("停止后不应再执行任务")
	}
	t.Log("停止后任务不再执行")
}
