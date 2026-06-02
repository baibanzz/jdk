package core

import "testing"

func TestName(t *testing.T) {
	l := NewLog(DefaultOption)
	l.Debug("ddd")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")

	l2 := l.Clone("demo")
	l2.Debug("ddd")
	l2.Info("info")
	l2.Warn("warn")
	l2.Error("error")

	l.Debug("ddd")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
}
