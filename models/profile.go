package models

import (
	"fmt"
)

//Timerange model for schedule
type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

//Schedule model
type DaySchedule struct {
	Active bool `json:"active"`
	Ranges []TimeRange `json:"ranges"`
}

//Profile model
type Profile struct {
	//Basic employee info
	ID int64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	MiddleInitial string `json:"middle_initial"`
	EmployeeID string `json:"employee_id"`

	// Job details
	Type EmployeeType `json:"type"`
	Title string `json:"title"`
	Department string `json:"department"`
	Rate float64 `json:"rate"`

	// Accounting codes 
	Fund string `json:"fund"`
	Org string `json:"org"`
	Acct string `json:"acct"`
	Prog string `json:"prog"`

	//Schedule map
	Schedule map[int]DaySchedule `json:"schedule"`
}

// Total hours calculation
func (ds DaySchedule) TotalHours() float64 {
	if !ds.Active {
		return 0.0
	}

	var total float64
	for _, r := range ds.Ranges {
		var startHour, startMin, endHour, endMin int
		fmt.Sscanf(r.Start, "%02d:%02d", &startHour, &startMin)
		fmt.Sscanf(r.End, "%02d:%02d", &endHour, &endMin)

		startTotal := float64(startHour) + float64(startMin)/60.0
		endTotal := float64(endHour) + float64(endMin)/60.0

		total += endTotal - startTotal
	}
	return total
}