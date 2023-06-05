package time

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

type Time time.Time

const timeFormat = "2006-01-02 15:04:05.000000"

func (t Time) MarshalJSON() ([]byte, error) {
	data := make([]byte, 0, len(timeFormat)+2)
	data = append(data, '"')
	data = time.Time(t).AppendFormat(data, timeFormat)
	data = append(data, '"')
	return data, nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.Local
	}
	parseTime, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), location)
	if err != nil {
		return err
	}
	*t = Time(parseTime)
	return nil
}

func (t Time) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t *Time) Scan(data any) error {
	value, ok := data.(time.Time)
	if !ok {
		return errors.New(fmt.Sprintf("%v不能被转换为Time", data))
	}
	*t = Time(value)
	return nil
}

func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	var timeData = time.Time(t)
	if timeData.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return timeData, nil
}

func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}
