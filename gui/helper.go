package gui

import (
	"strings"
	"time"
)

//Calculate hours for range lists
func CalculateHours(scheduleStr string) float64 {
	if scheduleStr == "" {
		return 0
	}

	//Initialize total hours
	var total float64
	parts := strings.Split(scheduleStr, ",")
	for _, part := range parts {
		//Check for valid range
		times := strings.Split(strings.TrimSpace(part), "-")
		if len(times) !=2{
			continue
		}

		//Parse start and end times
		start, _ := time.Parse("15:04", strings.TrimSpace(times[0]))
		end, _ := time.Parse("15:04", strings.TrimSpace(times[1]))

		//Calculate duration
		duration := end.Sub(start).Hours()
		if duration > 0 {
			total += duration
		}
	}

	return total
}