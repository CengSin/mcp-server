package util

import (
	"time"
)

func ParseTime(dateTime string) (*time.Time, error) {
	t, err := time.Parse(time.DateTime, dateTime)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
