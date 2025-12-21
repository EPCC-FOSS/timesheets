package models

type DailyEntry struct {
	Date string `json:"date"`

	//Base time
	HoursWorked float64 `json:"hours_worked"`

	// Full time specific
	SickLeave     float64 `json:"sick_leave"`
	Vacation      float64 `json:"vacation"`
	Holiday       float64 `json:"holiday"`
	CompTimeTaken float64 `json:"comp_time_taken"`
	OtherPaid     float64 `json:"other_paid"`
}

// TimesheetEntry model
type Timesheet struct {
	ID    int64 `json:"id"`
	Month int   `json:"month"`
	Year  int   `json:"year"`

	// Entries stored as json blob in DB. Marshal/Unmarshal needed later.
	Entries map[string]DailyEntry `json:"entries"`

	// Summary Data (Calculated on save)
	TotalWorked    float64 `json:"total_worked"`
	TotalOvertime  float64 `json:"total_overtime"`
	CompTimeEarned float64 `json:"comp_time_earned"`
}
