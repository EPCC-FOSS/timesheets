package pdfgen

import (
	"calendar_utility_node_for_timesheets/models"
	"fmt"
)

// GenerateTimesheet is the main entry point called by UI
func GenerateTimesheet(p *models.Profile, ts *models.Timesheet, outputPath string) error {
	// Route to appropriate generator based on employee type
	switch p.Type {
	case models.TypePartTime:
		return GeneratePartTimeTimesheet(p, ts, outputPath)
	case models.TypeFullTime:
		// TODO: Implement full-time timesheet generation
		return fmt.Errorf("full-time timesheet generation not yet implemented")
	case models.TypeWorkStudy:
		// TODO: Implement work-study timesheet generation
		return fmt.Errorf("work-study timesheet generation not yet implemented")
	default:
		return fmt.Errorf("unknown employee type: %v", p.Type)
	}
}
