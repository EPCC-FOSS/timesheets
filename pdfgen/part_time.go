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
	mrt.AddRow(5,
		col.New(6).Add(
			text.New(fmt.Sprintf("Name: %s %s %s", p.FirstName, p.MiddleInitial, p.LastName), props.Text{
				Size: 9,
			}),
		),
		col.New(3).Add(
			text.New(fmt.Sprintf("ID: %s", p.EmployeeID), props.Text{
				Size: 9,
			}),
		),
		col.New(3).Add(
			text.New(fmt.Sprintf("Position: %s", p.PositionNum), props.Text{
				Size: 9,
			}),
		),
	)

	mrt.AddRow(5,
		col.New(6).Add(
			text.New(fmt.Sprintf("Department: %s", p.Department), props.Text{
				Size: 9,
			}),
		),
		col.New(6).Add(
			text.New(fmt.Sprintf("Title: %s", p.Title), props.Text{
				Size: 9,
			}),
		),
	)

	mrt.AddRow(2, line.NewCol(12))
}

func addPartTimeTable(mrt core.Maroto, p *models.Profile, ts *models.Timesheet) {
	// Table header
	mrt.AddRow(6,
		col.New(2).Add(text.New("Date", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New("Day", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New("Hours Worked", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New("Overtime", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New("Total", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New("Weekly Total", props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
	)

	mrt.AddRow(1, line.NewCol(12))

	// Get sorted dates
	year, month := ts.Year, ts.Month
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()

	var weeklyRegular, weeklyOT float64

	threshold := p.Type.OvertimeThreshold()

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")
		dayName := date.Format("Mon")
		weekday := int(date.Weekday())

		// Track week start (Monday = 1)
		if weekday == 1 || day == 1 {
			weeklyRegular = 0
			weeklyOT = 0
		}

		entry, exists := ts.Entries[dateStr]
		var hoursWorked, overtime, total float64

		if exists {
			hoursWorked = entry.HoursWorked
			weeklyRegular += hoursWorked
		}

		// Show weekly total on Sunday or last day
		weeklyTotalStr := ""
		if weekday == 0 || day == daysInMonth {
			// Calculate overtime for the week
			if weeklyRegular > threshold {
				weeklyOT = weeklyRegular - threshold
				weeklyRegular = threshold
			}
			weeklyTotalStr = fmt.Sprintf("%.2f", weeklyRegular+weeklyOT)
		}

		total = hoursWorked + overtime

		mrt.AddRow(5,
			col.New(2).Add(text.New(date.Format("01/02/2006"), props.Text{Size: 8, Align: align.Center})),
			col.New(2).Add(text.New(dayName, props.Text{Size: 8, Align: align.Center})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", hoursWorked), props.Text{Size: 8, Align: align.Center})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", overtime), props.Text{Size: 8, Align: align.Center})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", total), props.Text{Size: 8, Align: align.Center})),
			col.New(2).Add(text.New(weeklyTotalStr, props.Text{Size: 8, Style: fontstyle.Bold, Align: align.Center})),
		)
	}

	mrt.AddRow(2, line.NewCol(12))

	// Calculate monthly totals
	var monthlyRegular, monthlyOT float64
	currentWeek := []models.DailyEntry{}
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	startOffset := int(firstOfMonth.Weekday()) - 1
	if startOffset < 0 {
		startOffset = 6
	}

	for i := 0; i < startOffset; i++ {
		currentWeek = append(currentWeek, models.DailyEntry{})
	}

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		var entry models.DailyEntry
		if e, ok := ts.Entries[dateStr]; ok {
			entry = e
		}

		currentWeek = append(currentWeek, entry)

		if len(currentWeek) == 7 {
			var weekTotal float64
			for _, e := range currentWeek {
				weekTotal += e.HoursWorked
			}
			if weekTotal > threshold {
				monthlyOT += weekTotal - threshold
				monthlyRegular += threshold
			} else {
				monthlyRegular += weekTotal
			}
			currentWeek = []models.DailyEntry{}
		}
	}

	// Handle partial week
	if len(currentWeek) > 0 {
		var weekTotal float64
		for _, e := range currentWeek {
			weekTotal += e.HoursWorked
		}
		if weekTotal > threshold {
			monthlyOT += weekTotal - threshold
			monthlyRegular += threshold
		} else {
			monthlyRegular += weekTotal
		}
	}

	monthlyTotal := monthlyRegular + monthlyOT

	// Monthly totals row
	mrt.AddRow(6,
		col.New(6).Add(text.New("MONTHLY TOTALS:", props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right})),
		col.New(2).Add(text.New(fmt.Sprintf("%.2f", monthlyRegular), props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New(fmt.Sprintf("%.2f", monthlyOT), props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New(fmt.Sprintf("%.2f", monthlyTotal), props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Center})),
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

	mrt.AddRow(8,
		col.New(6).Add(
			text.New("_________________________________", props.Text{Size: 9, Top: 6}),
			text.New("Employee Signature", props.Text{Size: 8}),
		),
		col.New(6).Add(
			text.New("_________________________________", props.Text{Size: 9, Top: 6}),
			text.New("Date", props.Text{Size: 8}),
		),
	)

	mrt.AddRow(3)

	mrt.AddRow(8,
		col.New(6).Add(
			text.New("_________________________________", props.Text{Size: 9, Top: 6}),
			text.New(fmt.Sprintf("Supervisor: %s", p.SupervisorName), props.Text{Size: 8}),
		),
		col.New(6).Add(
			text.New("_________________________________", props.Text{Size: 9, Top: 6}),
			text.New("Date", props.Text{Size: 8}),
		),
	)

	mrt.AddRow(5,
		col.New(12).Add(
			text.New(fmt.Sprintf("Supervisor Phone: %s", p.SupervisorPhone), props.Text{Size: 8}),
		),
	)
}
