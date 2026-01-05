package models

type DailyEntry struct {
	Date string `json:"date"`

	//Base time
	HoursWorked float64 `json:"hours_worked"`

	// Full time specific
	SickLeave     float64 `json:"sick_leave,omitempty"`
	Vacation      float64 `json:"vacation,omitempty"`
	Holiday       float64 `json:"holiday,omitempty"`
	CompTimeTaken float64 `json:"comp_time_taken,omitempty"`
	OtherPaid     float64 `json:"other_paid,omitempty"`

	// Overtime tracking (for part-time and full-time)
	OvertimeHours float64 `json:"overtime_hours,omitempty"`
}

// WeeklyEntry represents a week's worth of time tracking
type WeeklyEntry struct {
	WeekStartDate string                `json:"week_start_date"`
	WeekEndDate   string                `json:"week_end_date"`
	Days          map[string]DailyEntry `json:"days"` // Key: date string
	RegularTotal  float64               `json:"regular_total"`
	OvertimeTotal float64               `json:"overtime_total,omitempty"`
}

// TimesheetEntry model
type Timesheet struct {
	ID        int64 `json:"id"`
	ProfileID int64 `json:"profile_id"` // Link to profile
	Month     int   `json:"month"`
	Year      int   `json:"year"`

	// Entries stored as json blob in DB. Marshal/Unmarshal needed later.
	Entries map[string]DailyEntry `json:"entries"` // Kept for backward compatibility

	// Weekly entries (for better organization)
	Weeks []WeeklyEntry `json:"weeks,omitempty"`

	// Summary Data (Calculated on save)
	TotalWorked    float64 `json:"total_worked"`
	TotalOvertime  float64 `json:"total_overtime"`
	CompTimeEarned float64 `json:"comp_time_earned,omitempty"` // Full-time only

	// Full-time specific: Other Paid description for the week
	OtherPaidDescription string `json:"other_paid_description,omitempty"`

	// Work-Study specific
	GrossEarnings  float64 `json:"gross_earnings,omitempty"`  // Monthly gross earnings
	CurrentBalance float64 `json:"current_balance,omitempty"` // Hours used this month
	NewBalance     float64 `json:"new_balance,omitempty"`     // Remaining balance after this month
}
