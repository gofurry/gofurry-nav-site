package models

/*
 * @Desc: 统一时间接口
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"database/sql/driver"
	"fmt"
	"github.com/gofurry/gofurry-nav-backend/common"
	"time"
)

type LocalTime time.Time

const (
	timeFormat = common.TIME_FORMAT_DATE
	zone       = "Asia/Shanghai"
)

func (t *LocalTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = LocalTime(now)
	return
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	if &t == nil || t.IsZero() {
		return []byte("null"), nil
	}
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *LocalTime) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t LocalTime) String() string { return time.Time(t).Format(timeFormat) }

func (t LocalTime) Local() time.Time {
	loc, _ := time.LoadLocation(zone)
	return time.Time(t).In(loc)
}

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(t)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
