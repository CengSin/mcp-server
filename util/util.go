package util

import (
	"time"
)

var (
	Loc, _ = time.LoadLocation("Asia/Shanghai")
)

func ParseTime(dateTime string) (*time.Time, error) {
	t, err := time.ParseInLocation(time.DateTime, dateTime, Loc)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
