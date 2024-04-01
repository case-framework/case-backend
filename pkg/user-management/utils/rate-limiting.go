package utils

import "time"

func HasMoreAttemptsRecently(attempts []int64, moreThan int, intervalSeconds int64) bool {
	counter := 0
	startTime := time.Now().Unix() - intervalSeconds
	for _, ts := range attempts {
		if ts > startTime {
			counter += 1
		}
	}
	return counter > moreThan
}

func RemoveAttemptsOlderThan(attempts []int64, olderThanSeconds int64) []int64 {
	updated := []int64{}
	threshold := time.Now().Unix() - olderThanSeconds
	for _, v := range attempts {
		if v >= threshold {
			updated = append(updated, v)
		}
	}
	return updated
}
