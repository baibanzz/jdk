package model

import (
	"database/sql/driver"
	"time"
)

const ChinaLocation = "Asia/Shanghai"

// Time 中国时区时间类型，用于数据库读写
type Time struct {
	time.Time
}

// Value 实现 driver.Valuer 接口，写入数据库时转换为中国时区，MySQL 会存储为 datetime
func (t Time) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	// 转换为中国时区时间后写入数据库，time.Time 类型在 MySQL 中存储为 datetime
	loc, _ := time.LoadLocation(ChinaLocation)
	return t.Time.In(loc), nil
}

// Scan 实现 sql.Scanner 接口，从数据库读取时转换为中国时区
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		*t = Time{}
		return nil
	}

	var tTime time.Time
	loc, _ := time.LoadLocation(ChinaLocation)

	switch v := value.(type) {
	case time.Time:
		tTime = v
	case []byte:
		var err error
		tTime, err = time.ParseInLocation("2006-01-02 15:04:05", string(v), loc)
		if err != nil {
			return err
		}
	case string:
		var err error
		tTime, err = time.ParseInLocation("2006-01-02 15:04:05", v, loc)
		if err != nil {
			return err
		}
	default:
		return nil
	}

	t.Time = tTime.In(loc)
	return nil
}

func (t Time) String() string {
	return t.Time.Format("2006-01-02 15:04:05")
}

func (t Time) ToTime() time.Time {
	return t.Time
}

// Now 获取当前中国时间
func Now() Time {
	loc, _ := time.LoadLocation(ChinaLocation)
	return Time{time.Now().In(loc)}
}
