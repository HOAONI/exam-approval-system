package controllers

import (
	"time"
)

// parseTime 解析时间字符串
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}
