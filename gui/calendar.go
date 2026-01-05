package gui

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type CalendarPage struct {
	Repo   *db.Repository
	Window fyne.Window

	// State
	CurrentDate time.Time
	Profile     *models.Profile
	ShowDetails bool

	// UI components
	MonthLabel           *widget.Label
	WeeksContainer       *fyne.Container
	FooterLabel          *widget.Label
	MonthlyRegularLabel  *widget.Label
	MonthlyOvertimeLabel *widget.Label
	MonthlyTotalLabel    *widget.Label
	ToggleBtn            *widget.Button

	// Data Management
	DayWidgets            map[string]*DayCell
	WeeklyStatsContainers []fyne.CanvasObject
}

const StatsColumnWidth = 200 //Fixed column width for stats panel

func NewCalendarPage(win fyne.Window, repo *db.Repository) *CalendarPage {
	c := &CalendarPage{
		Repo:        repo,
		Window:      win,
		CurrentDate: time.Now(),
		ShowDetails: false,
		DayWidgets:  make(map[string]*DayCell),
	}

	c.MonthLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.WeeksContainer = container.NewVBox()
	c.FooterLabel = widget.NewLabelWithStyle("Loading...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.MonthlyRegularLabel = widget.NewLabelWithStyle("0.00", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.MonthlyOvertimeLabel = widget.NewLabelWithStyle("0.00", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.MonthlyTotalLabel = widget.NewLabelWithStyle("0.00", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	c.ToggleBtn = widget.NewButtonWithIcon("Show Extra Fields", theme.MenuDropDownIcon(), func() {
		c.ShowDetails = !c.ShowDetails
		for _, cell := range c.DayWidgets {
			cell.SetExtrasVisible(c.ShowDetails)
		}
	})
	c.ToggleBtn.Hide()

	return c
}

func (c *CalendarPage) BuildUI() fyne.CanvasObject {
	c.updateMonthLabel()

	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		c.CurrentDate = c.CurrentDate.AddDate(0, -1, 0)
		c.Refresh()
	})
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		c.CurrentDate = c.CurrentDate.AddDate(0, 1, 0)
		c.Refresh()
	})

	saveBtn := widget.NewButtonWithIcon("Save Changes", theme.DocumentSaveIcon(), c.saveData)
	exportBtn := widget.NewButtonWithIcon("Export to PDF", theme.DocumentIcon(), c.exportData)

	mainHeader := container.NewHBox(
		prevBtn, c.MonthLabel, nextBtn,
		layoutSpacer(0),
		c.ToggleBtn, saveBtn, exportBtn,
	)

	c.Refresh()

	// Create teal boxes for monthly summary metrics
	regularBox := c.createMetricBox("Regular Hours", c.MonthlyRegularLabel)
	overtimeBox := c.createMetricBox("Overtime Hours", c.MonthlyOvertimeLabel)
	totalBox := c.createMetricBox("Total Hours", c.MonthlyTotalLabel)

	footerContainer := container.NewGridWithColumns(3,
		regularBox,
		overtimeBox,
		totalBox,
	)

	return container.NewBorder(
		mainHeader,
		nil, nil, nil,
		container.NewBorder(
			c.buildWeekHeader(),
			container.NewPadded(footerContainer),
			nil, nil,
			container.NewScroll(c.WeeksContainer),
		),
	)
}

func (c *CalendarPage) buildWeekHeader() fyne.CanvasObject {
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dayGrid := container.NewGridWithColumns(7)
	for i := 0; i < 7; i++ {
		dayGrid.Add(widget.NewLabelWithStyle(days[i], fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	// Build stats header matching the table structure (3 columns)
	headerBreakdown := widget.NewLabelWithStyle("Breakdown", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerRegular := widget.NewLabelWithStyle("Regular", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerOvertime := widget.NewLabelWithStyle("Overtime", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	statsHeaderGrid := container.NewGridWithColumns(3, headerBreakdown, headerRegular, headerOvertime)
	statsHeaderCard := widget.NewCard("", "", statsHeaderGrid)
	rightContainer := c.makeFixedContainer(statsHeaderCard)

	return container.NewBorder(nil, nil, nil, rightContainer, dayGrid)
}

func (c *CalendarPage) Refresh() {
	c.updateMonthLabel()
	c.WeeksContainer.Objects = nil
	c.DayWidgets = make(map[string]*DayCell)
	c.WeeklyStatsContainers = nil

	prof, err := c.Repo.GetProfile()
	if err != nil {
		dialog.ShowError(err, c.Window)
		return
	}

	log.Printf("DEBUG: Attempting to load timesheet for Month=%d Year=%d\n", int(c.CurrentDate.Month()), c.CurrentDate.Year())

	//No profile set
	if prof == nil {
		return
	}

	c.Profile = prof

	if c.Profile.Type == models.TypeFullTime {
		c.ToggleBtn.Show()
	} else {
		c.ToggleBtn.Hide()
		c.ShowDetails = false
	}

	existingSheet, err := c.Repo.GetTimesheetByDate(int(c.CurrentDate.Month()), c.CurrentDate.Year())

	if err != nil {
		log.Println("DEBUG: Critical DB Error:", err)
	} else if existingSheet == nil {
		log.Println("DEBUG: No saved data found. Loading defaults")
	} else {
		log.Printf("DEBUG: Loaded Timesheet, found %d entries", len(existingSheet.Entries))
	}

	// Date Math
	year, month, _ := c.CurrentDate.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	var currentWeekCells []fyne.CanvasObject
	var currentWeekData []models.DailyEntry
	var grandTotalWork float64

	// Padding
	startOffset := int(firstOfMonth.Weekday()) - 1
	if startOffset < 0 {
		startOffset = 6
	}

	for i := 0; i < startOffset; i++ {
		currentWeekCells = append(currentWeekCells, layoutSpacer(10))
		currentWeekData = append(currentWeekData, models.DailyEntry{})
	}

	// Optimization: Define callback once
	onInputChanged := func() { c.recalculateLive() }

	// Render loop
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		var entry models.DailyEntry
		if existingSheet != nil && len(existingSheet.Entries) > 0 {
			if val, ok := existingSheet.Entries[dateStr]; ok {
				entry = val
			}
		} else {
			// Schedule Auto-fill
			dayOfWeek := int(date.Weekday()) - 1
			if dayOfWeek < 0 {
				dayOfWeek = 6
			}
			if sched, ok := c.Profile.Schedule[dayOfWeek]; ok && sched.Active {
				for _, r := range sched.Ranges {
					entry.HoursWorked += CalculateDailyHours(r.Start + "-" + r.End)
				}
			}
		}
		entry.Date = dateStr

		// Create widget
		cell := NewDayCell(day, entry, c.Profile.Type, onInputChanged)
		cell.SetExtrasVisible(c.ShowDetails)
		c.DayWidgets[dateStr] = cell

		currentWeekCells = append(currentWeekCells, cell.CanvasObj)
		currentWeekData = append(currentWeekData, entry)
		grandTotalWork += entry.HoursWorked

		if len(currentWeekCells) == 7 {
			c.renderWeekRow(currentWeekCells, currentWeekData)
			currentWeekCells = nil
			currentWeekData = nil
		}
	}

	// Final Padding
	if len(currentWeekCells) > 0 {
		for len(currentWeekCells) < 7 {
			currentWeekCells = append(currentWeekCells, layoutSpacer(10))
			currentWeekData = append(currentWeekData, models.DailyEntry{})
		}
		c.renderWeekRow(currentWeekCells, currentWeekData)
	}

	c.WeeksContainer.Refresh()
	// Calculate monthly totals for footer
	var monthlyGrandTotal, monthlyOT float64

	// Calculate weekly overtime and sum it up
	year, month, _ = c.CurrentDate.Date()
	firstOfMonth = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth = time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	startOffset = int(firstOfMonth.Weekday()) - 1
	if startOffset < 0 {
		startOffset = 6
	}

	var currentWeekDates []string
	// Add padding days
	for i := 0; i < startOffset; i++ {
		currentWeekDates = append(currentWeekDates, "")
	}

	// Process each day
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")
		currentWeekDates = append(currentWeekDates, dateStr)

		// When we have a full week, calculate overtime for that week
		if len(currentWeekDates) == 7 {
			var weekTotal float64
			for _, ds := range currentWeekDates {
				if ds != "" {
					if cell, ok := c.DayWidgets[ds]; ok {
						data := cell.GetData()
						dayTotal := data.HoursWorked + data.SickLeave + data.Vacation + data.Holiday + data.CompTimeTaken + data.OtherPaid
						weekTotal += dayTotal
						monthlyGrandTotal += dayTotal
					}
				}
			}

			threshold := c.Profile.Type.OvertimeThreshold()
			if weekTotal > threshold {
				monthlyOT += weekTotal - threshold
			}

			currentWeekDates = nil
		}
	}

	// Handle partial week at end of month
	if len(currentWeekDates) > 0 {
		var weekTotal float64
		for _, ds := range currentWeekDates {
			if ds != "" {
				if cell, ok := c.DayWidgets[ds]; ok {
					data := cell.GetData()
					dayTotal := data.HoursWorked + data.SickLeave + data.Vacation + data.Holiday + data.CompTimeTaken + data.OtherPaid
					weekTotal += dayTotal
					monthlyGrandTotal += dayTotal
				}
			}
		}

		threshold := c.Profile.Type.OvertimeThreshold()
		if weekTotal > threshold {
			monthlyOT += weekTotal - threshold
		}
	}

	monthlyRegular := monthlyGrandTotal - monthlyOT

	c.MonthlyRegularLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyRegular))
	c.MonthlyOvertimeLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyOT))
	c.MonthlyTotalLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyGrandTotal))
}

// createMetricBox creates a compact purple box with rounded corners for a metric
func (c *CalendarPage) createMetricBox(title string, valueLabel *widget.Label) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		titleLabel,
		valueLabel,
	)

	// Create a rounded rectangle background with purple color
	bg := canvas.NewRectangle(color.NRGBA{R: 156, G: 39, B: 176, A: 255}) // Purple color
	bg.CornerRadius = 10
	bg.SetMinSize(fyne.NewSize(140, 60))

	// Use smaller padding
	paddedContent := container.NewPadded(content)

	return container.NewStack(bg, paddedContent)
}

// renderWeekRow builds the row and calculates initial stats immediately
func (c *CalendarPage) renderWeekRow(cells []fyne.CanvasObject, data []models.DailyEntry) {
	dayGrid := container.NewGridWithColumns(7)
	for _, obj := range cells {
		dayGrid.Add(obj)
	}

	// Calculate stats and create table
	statsTable := generateWeeklyStatsTable(data, c.Profile.Type)
	c.WeeklyStatsContainers = append(c.WeeklyStatsContainers, statsTable)

	rightContainer := c.makeFixedContainer(statsTable)
	row := container.NewBorder(nil, nil, nil, rightContainer, dayGrid)

	c.WeeksContainer.Add(row)
	c.WeeksContainer.Add(widget.NewSeparator())
}

// recalculateLive scrapes UI widgets and updates text
func (c *CalendarPage) recalculateLive() {
	year, month, _ := c.CurrentDate.Date()
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

	startOffset := int(firstOfMonth.Weekday()) - 1
	if startOffset < 0 {
		startOffset = 6
	}

	var currentWeekData []models.DailyEntry
	weekIndex := 0
	var grandTotalWork float64

	for i := 0; i < startOffset; i++ {
		currentWeekData = append(currentWeekData, models.DailyEntry{})
	}

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		var d models.DailyEntry
		if cell, ok := c.DayWidgets[dateStr]; ok {
			d = cell.GetData()
		}

		currentWeekData = append(currentWeekData, d)
		grandTotalWork += d.HoursWorked

		// FIX: Strictly check for full week to increment
		if len(currentWeekData) == 7 {
			if weekIndex < len(c.WeeklyStatsContainers) {
				// Find the parent row container and update it
				newStatsTable := generateWeeklyStatsTable(currentWeekData, c.Profile.Type)
				// Replace the container's content
				if card, ok := c.WeeklyStatsContainers[weekIndex].(*widget.Card); ok {
					card.SetContent(newStatsTable.(*widget.Card).Content)
				}
			}
			currentWeekData = nil
			weekIndex++
		}
	}

	// Final partial week
	if len(currentWeekData) > 0 && weekIndex < len(c.WeeklyStatsContainers) {
		newStatsTable := generateWeeklyStatsTable(currentWeekData, c.Profile.Type)
		if card, ok := c.WeeklyStatsContainers[weekIndex].(*widget.Card); ok {
			card.SetContent(newStatsTable.(*widget.Card).Content)
		}
	}

	// Calculate monthly totals for footer - overtime calculated per week
	var monthlyGrandTotal, monthlyOT float64

	// Reset and recalculate weekly overtime
	currentWeekData = nil
	for i := 0; i < startOffset; i++ {
		currentWeekData = append(currentWeekData, models.DailyEntry{})
	}

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		var d models.DailyEntry
		if cell, ok := c.DayWidgets[dateStr]; ok {
			d = cell.GetData()
		}

		currentWeekData = append(currentWeekData, d)

		// When we have a full week, calculate overtime for that week
		if len(currentWeekData) == 7 {
			var weekTotal float64
			for _, entry := range currentWeekData {
				dayTotal := entry.HoursWorked + entry.SickLeave + entry.Vacation + entry.Holiday + entry.CompTimeTaken + entry.OtherPaid
				weekTotal += dayTotal
				monthlyGrandTotal += dayTotal
			}

			threshold := c.Profile.Type.OvertimeThreshold()
			if weekTotal > threshold {
				monthlyOT += weekTotal - threshold
			}

			currentWeekData = nil
		}
	}

	// Handle partial week at end of month
	if len(currentWeekData) > 0 {
		var weekTotal float64
		for _, entry := range currentWeekData {
			dayTotal := entry.HoursWorked + entry.SickLeave + entry.Vacation + entry.Holiday + entry.CompTimeTaken + entry.OtherPaid
			weekTotal += dayTotal
			monthlyGrandTotal += dayTotal
		}

		threshold := c.Profile.Type.OvertimeThreshold()
		if weekTotal > threshold {
			monthlyOT += weekTotal - threshold
		}
	}

	monthlyRegular := monthlyGrandTotal - monthlyOT

	c.MonthlyRegularLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyRegular))
	c.MonthlyOvertimeLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyOT))
	c.MonthlyTotalLabel.SetText(fmt.Sprintf("%.2f hrs", monthlyGrandTotal))
}

// Shared logic for stats table generation
func generateWeeklyStatsTable(data []models.DailyEntry, empType models.EmployeeType) fyne.CanvasObject {
	var tWork, tSick, tVac, tHol, tComp, tOther float64
	for _, d := range data {
		tWork += d.HoursWorked
		tSick += d.SickLeave
		tVac += d.Vacation
		tHol += d.Holiday
		tComp += d.CompTimeTaken
		tOther += d.OtherPaid
	}

	totalCompensated := tWork + tSick + tVac + tHol + tComp + tOther

	threshold := empType.OvertimeThreshold()
	var otHours, regularHours float64
	if totalCompensated > threshold {
		otHours = totalCompensated - threshold
		regularHours = threshold
	} else {
		regularHours = totalCompensated
	}

	// Create table header
	headerBreakdown := widget.NewLabelWithStyle("Breakdown", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerRegular := widget.NewLabelWithStyle("Regular", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	headerOvertime := widget.NewLabelWithStyle("Overtime", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	rows := []fyne.CanvasObject{
		container.NewGridWithColumns(3, headerBreakdown, headerRegular, headerOvertime),
	}

	// Add breakdown rows for full-time
	if empType == models.TypeFullTime {
		if tWork > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Work"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tWork), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
		if tSick > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Sick Leave"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tSick), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
		if tVac > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Vacation"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tVac), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
		if tHol > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Holiday"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tHol), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
		if tComp > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Comp Time"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tComp), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
		if tOther > 0 {
			rows = append(rows, container.NewGridWithColumns(3,
				widget.NewLabel("Other Paid"),
				widget.NewLabelWithStyle(fmt.Sprintf("%.2f", tOther), fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
			))
		}
	}

	// Add totals row with separator
	rows = append(rows, widget.NewSeparator())
	rows = append(rows, container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Weekly Total", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(fmt.Sprintf("%.2f", regularHours), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(fmt.Sprintf("%.2f", otHours), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	))

	tableContent := container.NewVBox(rows...)
	return widget.NewCard("", "", tableContent)
}

// ... updateMonthLabel, saveData, exportData remain the same ...
func (c *CalendarPage) updateMonthLabel() {
	c.MonthLabel.SetText(c.CurrentDate.Format("January 2006"))
}

func (c *CalendarPage) saveData() {
	entries := make(map[string]models.DailyEntry)
	var totalWorked float64
	for dateStr, cell := range c.DayWidgets {
		data := cell.GetData()
		entries[dateStr] = data
		totalWorked += data.HoursWorked
	}

	log.Printf("DEBUG: Saving Timesheet -> Month: %d, Year: %d, Total Entries: %d, Total Hours: %.2f\n",
		int(c.CurrentDate.Month()), c.CurrentDate.Year(), len(entries), totalWorked)

	ts := models.Timesheet{
		Month:       int(c.CurrentDate.Month()),
		Year:        c.CurrentDate.Year(),
		Entries:     entries,
		TotalWorked: totalWorked,
	}
	if err := c.Repo.SaveTimesheet(ts); err != nil {
		log.Println("DEBUG: Save FAILED:", err)
		dialog.ShowError(err, c.Window)
		return
	}
	dialog.ShowInformation("Saved", "Timesheet Updated Successfully.", c.Window)
	log.Println("DEBUG: Save SUCCESS")
}

func (c *CalendarPage) makeFixedContainer(obj fyne.CanvasObject) fyne.CanvasObject {
	// Create transparent spacer
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(StatsColumnWidth, 10))

	// Use stack, spacer pushes width
	return container.NewStack(spacer, container.NewPadded(obj))
}

func (c *CalendarPage) exportData() {
	dialog.ShowInformation("Export", "Export functionality not implemented yet.", c.Window)
}
