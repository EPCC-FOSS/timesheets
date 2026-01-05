package models

import (
	"fmt"
)

// Timerange model for schedule
type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Schedule model
type DaySchedule struct {
	Active bool        `json:"active"`
	Ranges []TimeRange `json:"ranges"`
}

// AccountingCodes represents a set of accounting codes (for part-time employees who may have 2 rows)
type AccountingCodes struct {
	Fund         string  `json:"fund"`
	Organization string  `json:"organization"`
	Account      string  `json:"account"`
	Program      string  `json:"program"`
	HourlyRate   float64 `json:"hourly_rate,omitempty"` // For part-time second row
}

// Profile model
type Profile struct {
	//Basic employee info
	ID            int64  `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	MiddleName    string `json:"middle_name"`    // Full name for full-time
	MiddleInitial string `json:"middle_initial"` // Initial for part-time/work-study
	EmployeeID    string `json:"employee_id"`

	// Job details
	Type        EmployeeType `json:"type"`
	Title       string       `json:"title"`
	PositionNum string       `json:"position_num"` // Position Number
	Department  string       `json:"department"`
	Rate        float64      `json:"rate"`     // Hourly Rate (primary)
	Location    string       `json:"location"` // For full-time

	// Accounting codes - Primary (Row 1)
	PrimaryAccounting AccountingCodes `json:"primary_accounting"`

	// Secondary accounting codes (for part-time employees with 2 rows)
	SecondaryAccounting *AccountingCodes `json:"secondary_accounting,omitempty"`

	// Supervisor information
	SupervisorName  string `json:"supervisor_name"`  // Supervisor print name
	SupervisorPhone string `json:"supervisor_phone"` // Supervisor phone

	// Contact information
	EmployeePhone string `json:"employee_phone"` // Employee office phone
	OfficePhone   string `json:"office_phone"`   // Department/Office phone (different from employee phone)

	// Work-Study specific
	SemesterAllocation float64 `json:"semester_allocation,omitempty"` // Total hours allocated for the semester
	PreviousBalance    float64 `json:"previous_balance,omitempty"`    // Balance from previous timesheet

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
