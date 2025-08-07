/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package time

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

type Time time.Time

var timeFormatList = []string{
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func (t Time) MarshalJSON() ([]byte, error) {
	timeFormat := timeFormatList[0]
	data := make([]byte, 0, len(timeFormat)+2)
	data = append(data, '"')
	data = time.Time(t).AppendFormat(data, timeFormat)
	data = append(data, '"')
	return data, nil
}

func (t *Time) parse(data string, format string) error {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.Local
	}
	parseTime, err := time.ParseInLocation(format, data, location)
	if err != nil {
		return err
	}
	*t = Time(parseTime)
	return nil
}

func (t *Time) Parse(data string) error {
	errList := make([]error, 0, len(timeFormatList))
	for _, timeFormat := range timeFormatList {
		if err := t.parse(data, timeFormat); err == nil {
			return nil
		} else {
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}

func Parse(data string) (*Time, error) {
	errList := make([]error, 0, len(timeFormatList))
	for _, timeFormat := range timeFormatList {
		t := &Time{}
		if err := t.parse(data, timeFormat); err != nil {
			errList = append(errList, err)
		} else {
			return t, nil
		}
	}
	return nil, errors.Join(errList...)
}

func (t *Time) UnmarshalJSON(data []byte) error {
	errList := make([]error, 0, len(timeFormatList))
	for _, timeFormat := range timeFormatList {
		if err := t.parse(string(data), `"`+timeFormat+`"`); err != nil {
			errList = append(errList, err)
		} else {
			return nil
		}
	}
	return errors.Join(errList...)
}

func (t Time) String() string {
	return time.Time(t).Format(timeFormatList[0])
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
