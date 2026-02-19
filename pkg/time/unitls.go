package time

import (
	"time"
)

// ParseDateFromFrontend parses date like "11.11.2021" into *time.Time
func ParseDateFromFrontend(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	layout := "02.01.2006" // frontend format: 11.11.2021
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return nil // agar format noto‘g‘ri bo‘lsa, nil qaytaradi
	}
	return &t
}
