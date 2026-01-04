package gui

import (
	"strings"
	"time"
)

// Calculate hours for range lists
func CalculateDailyHours(scheduleStr string) float64 {
	if scheduleStr == "" {
		return 0
	}

	//Initialize total hours
	var total float64
	parts := strings.Split(scheduleStr, ",")
	for _, part := range parts {
		//Check for valid range
		times := strings.Split(strings.TrimSpace(part), "-")
		if len(times) != 2 {
			continue
		}

		startStr := strings.TrimSpace(times[0])
		endStr := strings.TrimSpace(times[1])

		//Parse start and end times
		start, err1 := parseFlexibleTime(startStr)
		end, err2 := parseFlexibleTime(endStr)

		if err1 == nil && err2 == nil {
			duration := end.Sub(start).Hours()
			//Handle negative range (eg 2:00 to 22:00)
			if duration > 0 {
				total += duration
			}
		}
	}

	return total
}

// Parsing helper to handle single and double difit time
func parseFlexibleTime(t string) (time.Time, error) {
	// Try 24 hour format
	if parsed, err := time.Parse("15:04", t); err == nil {
		return parsed, nil
	}

	//Try single digit
	if parsed, err := time.Parse("3:04", t); err == nil {
		return parsed, nil
	}

	// Manual padding if all else fails
	if len(t) == 4 && strings.Index(t, ":") == 1 {
		return time.Parse("15:04", "0"+t)
	}

	return time.Time{}, nil
}
