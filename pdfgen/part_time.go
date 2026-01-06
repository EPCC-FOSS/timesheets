package pdfgen

import (
	"calendar_utility_node_for_timesheets/models"
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// GeneratePartTimeTimesheet generates a PDF for part-time employees
func GeneratePartTimeTimesheet(p *models.Profile, ts *models.Timesheet, outputPath string) error {
	cfg := config.NewBuilder().
		WithDimensions(215.9, 279.4).
		WithLeftMargin(10).
		WithTopMargin(2).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)

	// Add header
	addPartTimeHeader(mrt, p, ts)

	// Add employee info
	addPartTimeEmployeeInfo(mrt, p)

	// Add timesheet table
	addPartTimeTable(mrt, p, ts)

	// Add accounting codes section
	addPartTimeAccounting(mrt, p)

	// Add signature section
	addPartTimeSignatures(mrt, p)

	// Save PDF
	doc, err := mrt.Generate()
	if err != nil {
		return err
	}

	return doc.Save(outputPath)
}

func addPartTimeHeader(mrt core.Maroto, p *models.Profile, ts *models.Timesheet) {
	// First row: Logo centered
	mrt.AddRow(28,
		image.NewFromFileCol(12, "assets/epcc_logo.png", props.Rect{
			Center:  true,
			Percent: 88,
		}),
	)

	// Second row: Title (pulled up with negative Top to reduce gap)
	mrt.AddRow(5,
		col.New(12).Add(
			text.New("PART-TIME NON-FACULTY TIMESHEET FOR", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   0,
			}),
		),
	)

	// Third row: Month and Year with underlines and labels
	monthName := time.Month(ts.Month).String()
	yearStr := fmt.Sprintf("%d", ts.Year)

	// Create underlined month and year by adding a line below
	mrt.AddRow(7,
		col.New(4),
		col.New(2).Add(
			text.New(monthName, props.Text{
				Size:  11,
				Align: align.Center,
			}),
		),
		col.New(2).Add(
			text.New(yearStr, props.Text{
				Size:  12,
				Align: align.Center,
			}),
		),
		col.New(4),
	)

	// Underlines
	mrt.AddRow(1,
		col.New(4),
		line.NewCol(2),
		line.NewCol(2),
		col.New(4),
	)

	mrt.AddRow(4,
		col.New(4),
		col.New(2).Add(
			text.New("Month", props.Text{
				Size:  9,
				Align: align.Center,
			}),
		),
		col.New(2).Add(
			text.New("Year", props.Text{
				Size:  9,
				Align: align.Center,
			}),
		),
		col.New(4),
	)
}

func addPartTimeEmployeeInfo(mrt core.Maroto, p *models.Profile) {
	// First row: Name and ID fields
	mrt.AddRow(6,
		col.New(3).Add(
			text.New("LAST NAME", props.Text{Size: 8, Style: fontstyle.Bold}),
		),
		col.New(3).Add(
			text.New("FIRST NAME", props.Text{Size: 8, Style: fontstyle.Bold}),
		),
		col.New(2).Add(
			text.New("MI", props.Text{Size: 8, Style: fontstyle.Bold}),
		),
		col.New(4).Add(
			text.New("EMPLOYEE ID", props.Text{Size: 8, Style: fontstyle.Bold}),
		),
	)

	mrt.AddRow(5,
		col.New(3).Add(
			text.New(p.LastName, props.Text{Size: 9}),
		),
		col.New(3).Add(
			text.New(p.FirstName, props.Text{Size: 9}),
		),
		col.New(2).Add(
			text.New(p.MiddleInitial, props.Text{Size: 9}),
		),
		col.New(4).Add(
			text.New(p.EmployeeID, props.Text{Size: 9}),
		),
	)

	// Second row: Department and Position
	mrt.AddRow(6,
		col.New(8).Add(
			text.New(fmt.Sprintf("DEPARTMENT: %s", p.Department), props.Text{Size: 9}),
		),
		col.New(4).Add(
			text.New(fmt.Sprintf("POSITION NO: %s", p.PositionNum), props.Text{Size: 9}),
		),
	)

	// Primary accounting row
	mrt.AddRow(6,
		col.New(2).Add(
			text.New(fmt.Sprintf("FUND: %s", p.PrimaryAccounting.Fund), props.Text{Size: 8}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("ORG: %s", p.PrimaryAccounting.Organization), props.Text{Size: 8}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("ACCT: %s", p.PrimaryAccounting.Account), props.Text{Size: 8}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("PROG: %s", p.PrimaryAccounting.Program), props.Text{Size: 8}),
		),
		col.New(4).Add(
			text.New(fmt.Sprintf("HOURLY RATE: $%.2f", p.Rate), props.Text{Size: 8}),
		),
	)

	// Secondary accounting row if exists
	if p.SecondaryAccounting != nil {
		mrt.AddRow(6,
			col.New(2).Add(
				text.New(fmt.Sprintf("FUND: %s", p.SecondaryAccounting.Fund), props.Text{Size: 8}),
			),
			col.New(2).Add(
				text.New(fmt.Sprintf("ORG: %s", p.SecondaryAccounting.Organization), props.Text{Size: 8}),
			),
			col.New(2).Add(
				text.New(fmt.Sprintf("ACCT: %s", p.SecondaryAccounting.Account), props.Text{Size: 8}),
			),
			col.New(2).Add(
				text.New(fmt.Sprintf("PROG: %s", p.SecondaryAccounting.Program), props.Text{Size: 8}),
			),
			col.New(4),
		)
	}
}

func addPartTimeTable(mrt core.Maroto, p *models.Profile, ts *models.Timesheet) {
	// Table header - "WEEK" and "NUMBER OF HOURS" sections
	mrt.AddRow(5,
		col.New(3).Add(
			text.New("WEEK", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center}),
		),
		col.New(9).Add(
			text.New("NUMBER OF HOURS", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center}),
		),
	)

	// Column headers
	mrt.AddRow(5,
		col.New(1).Add(text.New("FROM", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("TO", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("M", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("T", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("W", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("TH", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("F", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("S", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("S", props.Text{Size: 7, Align: align.Center})),
		col.New(2).Add(text.New("REG", props.Text{Size: 7, Align: align.Center})),
		col.New(1).Add(text.New("TOTAL HOURS", props.Text{Size: 6, Align: align.Center})),
	)

	mrt.AddRow(1, line.NewCol(12))

	// Get weeks data
	year, month := ts.Year, ts.Month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)

	// Find the Monday of the first week
	weekStart := firstDay
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}

	threshold := p.Type.OvertimeThreshold()
	var monthlyRegular, monthlyOT float64

	// Generate up to 5 weeks
	for weekNum := 0; weekNum < 5; weekNum++ {
		weekEnd := weekStart.AddDate(0, 0, 6) // Sunday

		// Collect hours for each day of the week
		dayHours := make([]float64, 7) // Mon-Sun
		var weekTotal float64

		for dayOffset := 0; dayOffset < 7; dayOffset++ {
			currentDay := weekStart.AddDate(0, 0, dayOffset)

			// Only count if within current month
			if currentDay.Month() == time.Month(month) && currentDay.Year() == year {
				dateStr := currentDay.Format("2006-01-02")
				if entry, exists := ts.Entries[dateStr]; exists {
					dayHours[dayOffset] = entry.HoursWorked
					weekTotal += entry.HoursWorked
				}
			}
		}

		// Calculate regular and OT for this week
		var weekRegular, weekOT float64
		if weekTotal > threshold {
			weekOT = weekTotal - threshold
			weekRegular = threshold
		} else {
			weekRegular = weekTotal
		}

		monthlyRegular += weekRegular
		monthlyOT += weekOT

		// Helper to format hours (empty string if zero)
		formatHours := func(hours float64) string {
			if hours == 0 {
				return ""
			}
			return fmt.Sprintf("%.2f", hours)
		}

		// Add week row
		mrt.AddRow(7,
			col.New(1).Add(text.New(weekStart.Format("01/02/06"), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(weekEnd.Format("01/02/06"), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[0]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[1]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[2]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[3]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[4]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[5]), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(dayHours[6]), props.Text{Size: 7, Align: align.Center})),
			col.New(2).Add(text.New(formatHours(weekRegular), props.Text{Size: 7, Align: align.Center})),
			col.New(1).Add(text.New(formatHours(weekTotal), props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		)

		// Move to next week
		weekStart = weekStart.AddDate(0, 0, 7)

		// Stop if we've gone past the month
		if weekStart.Month() != time.Month(month) && weekStart.Day() > 7 {
			break
		}
	}

	mrt.AddRow(2, line.NewCol(12))

	// Note about rounding
	mrt.AddRow(4,
		col.New(9).Add(
			text.New("Round off hours worked to the nearest quarter hour; ¼ hr = .25; ½ hr. = .50; ¾ hr. = .75; 1 hr. = 1", props.Text{Size: 7}),
		),
		col.New(3).Add(
			text.New("TOTAL HOURS", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Right}),
		),
	)

	monthlyTotal := monthlyRegular + monthlyOT

	// Monthly total
	mrt.AddRow(6,
		col.New(9),
		col.New(3).Add(
			text.New(fmt.Sprintf("%.2f", monthlyTotal), props.Text{Size: 12, Style: fontstyle.Bold, Align: align.Center}),
		),
	)
}

func addPartTimeAccounting(mrt core.Maroto, p *models.Profile) {
	mrt.AddRow(3)

	mrt.AddRow(5,
		col.New(12).Add(
			text.New("PRIMARY ACCOUNTING CODES", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
	)

	mrt.AddRow(5,
		col.New(3).Add(text.New(fmt.Sprintf("Fund: %s", p.PrimaryAccounting.Fund), props.Text{Size: 8})),
		col.New(3).Add(text.New(fmt.Sprintf("Org: %s", p.PrimaryAccounting.Organization), props.Text{Size: 8})),
		col.New(3).Add(text.New(fmt.Sprintf("Acct: %s", p.PrimaryAccounting.Account), props.Text{Size: 8})),
		col.New(3).Add(text.New(fmt.Sprintf("Prog: %s", p.PrimaryAccounting.Program), props.Text{Size: 8})),
	)

	// Secondary accounting if exists
	if p.SecondaryAccounting != nil {
		mrt.AddRow(5,
			col.New(12).Add(
				text.New("SECONDARY ACCOUNTING CODES", props.Text{
					Size:  9,
					Style: fontstyle.Bold,
				}),
			),
		)

		mrt.AddRow(5,
			col.New(3).Add(text.New(fmt.Sprintf("Fund: %s", p.SecondaryAccounting.Fund), props.Text{Size: 8})),
			col.New(3).Add(text.New(fmt.Sprintf("Org: %s", p.SecondaryAccounting.Organization), props.Text{Size: 8})),
			col.New(3).Add(text.New(fmt.Sprintf("Acct: %s", p.SecondaryAccounting.Account), props.Text{Size: 8})),
			col.New(3).Add(text.New(fmt.Sprintf("Prog: %s", p.SecondaryAccounting.Program), props.Text{Size: 8})),
		)
	}
}

func addPartTimeSignatures(mrt core.Maroto, p *models.Profile) {
	mrt.AddRow(3)

	mrt.AddRow(4,
		col.New(12).Add(
			text.New("I certify that the above time record is true and accurate.", props.Text{Size: 7}),
		),
	)

	mrt.AddRow(2)

	// Supervisor and Employee Signature labels
	mrt.AddRow(5,
		col.New(5).Add(text.New("Supervisor Signature:", props.Text{Size: 8})),
		col.New(2).Add(text.New("Date:", props.Text{Size: 8})),
		col.New(3).Add(text.New("Employee's Signature:", props.Text{Size: 8})),
		col.New(2).Add(text.New("Date:", props.Text{Size: 8})),
	)

	// Signature lines
	mrt.AddRow(1,
		line.NewCol(5),
		col.New(2),
		line.NewCol(3),
		col.New(2),
	)

	mrt.AddRow(1,
		col.New(5),
		line.NewCol(2),
		col.New(3),
		line.NewCol(2),
	)

	mrt.AddRow(2)

	// Supervisor Print Name
	mrt.AddRow(5,
		col.New(6).Add(text.New("Supervisor Print Name:", props.Text{Size: 8})),
	)

	mrt.AddRow(1,
		line.NewCol(6),
	)

	mrt.AddRow(2)

	// Phone numbers
	mrt.AddRow(5,
		col.New(6).Add(text.New("Supervisor Office Phone Number:", props.Text{Size: 8})),
		col.New(6).Add(text.New("Employee Office Phone Number:", props.Text{Size: 8})),
	)

	mrt.AddRow(1,
		line.NewCol(6),
		line.NewCol(6),
	)
}
