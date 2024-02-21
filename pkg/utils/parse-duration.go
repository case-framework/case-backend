package utils

import (
	"fmt"
	"time"
)

func ParseDurationString(value string) (time.Duration, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return time.Duration(0), fmt.Errorf("invalid time duration '%s' : %s", value, err.Error())
	}
	return d, nil
}
