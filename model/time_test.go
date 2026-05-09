package model

import (
	"database/sql/driver"
	"testing"
	"time"
)

// 确保 Time 类型实现了 driver.Valuer 和 sql.Scanner 接口
var _ driver.Valuer = Time{}
var _ interface{ Scan(src interface{}) error } = (*Time)(nil)

func TestTime_Value(t *testing.T) {
	// 测试零值
	zeroTime := Time{}
	val, err := zeroTime.Value()
	if err != nil {
		t.Errorf("Zero time Value() error: %v", err)
	}
	if val != nil {
		t.Errorf("Zero time Value() expected nil, got %v", val)
	}

	// 测试非零值
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	testTime := Time{now}

	val, err = testTime.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}

	// 验证返回值是中国时区
	resultTime, ok := val.(time.Time)
	if !ok {
		t.Errorf("Value() expected time.Time, got %T", val)
	}

	if resultTime.Location().String() != "Asia/Shanghai" {
		t.Errorf("Value() expected Asia/Shanghai, got %s", resultTime.Location().String())
	}
}

func TestTime_Scan(t *testing.T) {
	// 测试 nil 值
	t1 := &Time{}
	err := t1.Scan(nil)
	if err != nil {
		t.Errorf("Scan(nil) error: %v", err)
	}
	if !t1.Time.IsZero() {
		t.Errorf("Scan(nil) expected zero time")
	}

	// 测试 time.Time 类型
	loc, _ := time.LoadLocation("Asia/Shanghai")
	inputTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	t2 := &Time{}
	err = t2.Scan(inputTime)
	if err != nil {
		t.Errorf("Scan(time.Time) error: %v", err)
	}
	if t2.Time.Location().String() != "Asia/Shanghai" {
		t.Errorf("Scan(time.Time) expected Asia/Shanghai, got %s", t2.Time.Location().String())
	}

	// 测试 string 类型
	t3 := &Time{}
	err = t3.Scan("2024-01-15 10:30:00")
	if err != nil {
		t.Errorf("Scan(string) error: %v", err)
	}
	if t3.Time.Location().String() != "Asia/Shanghai" {
		t.Errorf("Scan(string) expected Asia/Shanghai, got %s", t3.Time.Location().String())
	}
	if t3.Time.Year() != 2024 || t3.Time.Month() != time.January || t3.Time.Day() != 15 {
		t.Errorf("Scan(string) expected 2024-01-15, got %s", t3.Time.Format("2006-01-02"))
	}

	// 测试 []byte 类型
	t4 := &Time{}
	err = t4.Scan([]byte("2024-01-15 10:30:00"))
	if err != nil {
		t.Errorf("Scan([]byte) error: %v", err)
	}
	if t4.Time.Location().String() != "Asia/Shanghai" {
		t.Errorf("Scan([]byte) expected Asia/Shanghai, got %s", t4.Time.Location().String())
	}
}

func TestTime_String(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	testTime := Time{time.Date(2024, 1, 15, 10, 30, 0, 0, loc)}

	result := testTime.String()
	expected := "2024-01-15 10:30:00"

	if result != expected {
		t.Errorf("String() expected %s, got %s", expected, result)
	}
}

func TestTime_ToTime(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	inputTime := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	testTime := Time{inputTime}

	result := testTime.ToTime()

	if !result.Equal(inputTime) {
		t.Errorf("ToTime() expected %v, got %v", inputTime, result)
	}
}

func TestNow(t *testing.T) {
	now := Now()

	// 验证是中国时区
	if now.Time.Location().String() != "Asia/Shanghai" {
		t.Errorf("Now() expected Asia/Shanghai, got %s", now.Time.Location().String())
	}

	// 验证时间在合理范围内（当前时刻前后1分钟）
	nowReal := time.Now().In(now.Time.Location())
	diff := nowReal.Sub(now.Time)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Now() time difference too large: %v", diff)
	}
}
