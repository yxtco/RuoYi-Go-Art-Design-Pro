package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const LocalTimeFormat = "2006-01-02 15:04:05"
const LocalDateFormat = "2006-01-02"

// LocalTime 自定义时间类型
type LocalTime struct {
	time.Time
}

// MarshalJSON 自定义 JSON 编码方法
func (ct LocalTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(ct.Time).Format(LocalTimeFormat))
	// null 值返回空
	if stamp == `"0001-01-01 00:00:00"` {
		return []byte(`""`), nil
	}
	return []byte(stamp), nil
}

// UnmarshalJSON 自定义 JSON 解码方法
func (ct *LocalTime) UnmarshalJSON(b []byte) error {
	// 将字符串解析为时间
	if string(b) != "" && string(b) != `""` {
		t, err := time.Parse(`"`+LocalTimeFormat+`"`, string(b))
		if err != nil {
			return err
		}
		ct.Time = t
	}
	return nil
}

func (ct LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if ct.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ct.Time, nil
}
func (ct *LocalTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*ct = LocalTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (ct LocalTime) String() string {
	return ct.Time.Format(LocalTimeFormat)
}

// LocalDate 自定义时间类型
type LocalDate struct {
	time.Time
}

// MarshalJSON 自定义 JSON 编码方法
func (ct LocalDate) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(ct.Time).Format(LocalDateFormat))
	// null 值返回空
	if stamp == `"0001-01-01 00:00:00"` {
		return []byte(`""`), nil
	}
	return []byte(stamp), nil
}

// UnmarshalJSON 自定义 JSON 解码方法
func (ct *LocalDate) UnmarshalJSON(b []byte) error {
	// 将字符串解析为时间
	if string(b) != "" && string(b) != `""` {
		t, err := time.Parse(`"`+LocalDateFormat+`"`, string(b))
		if err != nil {
			return err
		}
		ct.Time = t
	}
	return nil
}

func (ct LocalDate) Value() (driver.Value, error) {
	var zeroTime time.Time
	if ct.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ct.Time, nil
}
func (ct *LocalDate) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*ct = LocalDate{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
