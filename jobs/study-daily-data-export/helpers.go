package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// StartOfDay returns the start time of the given date (00:00:00)
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end time of the given date (23:59:59.999999999)
func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func responseFileName(date time.Time, surveyKey string, format string) string {
	dateStr := date.Format("2006-01-02")
	suffix := ""
	switch format {
	case "wide":
		suffix = "wide.csv"
	case "long":
		suffix = "long.csv"
	case "json":
		suffix = "json.json"
	}
	return fmt.Sprintf("%s##responses##%s##%s", dateStr, surveyKey, suffix)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		return false
	}
	return !info.IsDir()
}
