package timeutil

import (
	"fmt"
	"time"
)

const naiveDateTimeLayout = "2006-01-02T15:04:05"

func ParseUserTime(s string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		return time.Time{}, fmt.Errorf("parse user time: location is required")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation(naiveDateTimeLayout, s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("parse user time %q: expected RFC3339 with offset or naive datetime", s)
}
