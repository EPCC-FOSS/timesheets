package gui

import (
	"fmt"
	"time"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type CalendarPage struct {
	Repo   *db.Repository
	Window fyne.Window

	//State
	CurrentDate time.Time
	Profile     *models.Profile
	ShowDetails bool

	//UI components
	MonthLabel    *widget.Label
	WeeksContainer *fyne.Container
	FooterLabel *widget.Label
	ToggleBtn *widget.Button

	//Daywidgets
	DayWidgets map[string]*DayCell
}

// Ui for calendar page
func NewCalendarPage(win fyne.Window, repo *db.Repository) *CalendarPage {
	c := &CalendarPage{
		Repo:        repo,
		Window:      win,
		CurrentDate: time.Now(),
		ShowDetails: false,
		DayWidgets:  make(map[string]*DayCell),
	}

	//Widgets init
	c.MonthLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.WeeksContainer = container.NewVBox()
	c.FooterLabel = widget.NewLabel("Loading...")

	//Toggle fields for full time
	c.ToggleBtn = widget.NewButtonWithIcon("Show Extra Fields", theme.MenuDropDownIcon(), func ()  {
		c.ShowDetails = !c.ShowDetails //Flip state

		for _, cell := range c.DayWidgets {
			cell.SetExtrasVisible(c.ShowDetails)
		}
	})
	c.ToggleBtn.Hide()

	return c
}

func (c *CalendarPage) BuildUI() fyne.CanvasObject {
	c.updateMonthLabel()

	// Month Navigation
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		c.CurrentDate = c.CurrentDate.AddDate(0, -1., 0)
		c.Refresh()
	})
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		c.CurrentDate = c.CurrentDate.AddDate(0, 1, 0)
		c.Refresh()
	})

	//Save and export buttons
	saveBtn := widget.NewButtonWithIcon("Save Changes", theme.DocumentSaveIcon(), c.saveData)
	exportBtn := widget.NewButtonWithIcon("Export to PDF", theme.DocumentIcon(), c.exportData)

	mainHeader := container.NewHBox(
		prevBtn,
		c.MonthLabel,
		nextBtn,
		layoutSpacer(0),
		c.ToggleBtn,
		saveBtn,
		exportBtn,
	)

	// Calendar nesed laYOUT
	calendaContent := container.NewBorder(
		c.buildWeekHeader(),
		c.FooterLabel,
		nil, nil,
		container.NewScroll(c.WeeksContainer),
	)

	//Calendar tab built
	return container.NewBorder(
		mainHeader,
		nil, nil, nil,
		calendaContent,
	)
}

func (c *CalendarPage) buildWeekHeader() fyne.CanvasObject {
	// Weekday headers
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun", "Weekly Totals"}
	
	// Grid for days
	dayGrid := container.NewGridWithColumns(7)
	for i := 0; i < 7; i++ {
		dayGrid.Add(widget.NewLabelWithStyle(days[i], fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	// Weekly stats container
	statsLabel := widget.NewLabelWithStyle("Weekly Stats", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	// Split containing both weekly total stats and days
	split := container.NewHSplit(dayGrid, statsLabel)
	split.SetOffset(0.85)
	return split
}

func (c *CalendarPage) Refresh() {
	//Clear items
	c.updateMonthLabel()

	//Setup
	c.WeeksContainer.Objects = nil
	c.DayWidgets = make(map[string]*DayCell)

	// get proofile
	prof, err := c.Repo.GetProfile()
	if err != nil {
		dialog.ShowError(err, c.Window)
		return
	}
	c.Profile = prof

	// Include full time fields toggle if employee is full time
	if c.Profile.Type == models.TypeFullTime {
		c.ToggleBtn.Show()
	}else{
		c.ToggleBtn.Hide()
		c.ShowDetails = false
	}

	// Check if existing timesheet for month
	existingSheet, _ := c.Repo.GetTimesheetByDate(int(c.CurrentDate.Month()), c.CurrentDate.Year())

	//Calculate days
	year, month, _ := c.CurrentDate.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	// Buffer days into slice of len 7 to render week as a row
	var currentWeekCells []fyne.CanvasObject
	var currentWeekData []models.DailyEntry

	// Determine padding for first day (eg, 1st on monday,
	// 1 cell of padding to the left of the first row)
	startOffset := int(firstOfMonth.Weekday()) - 1
	if startOffset < 0 {startOffset = 6} //sunday fix

	//Add padding
	for i := 0; i < startOffset; i++ {
		currentWeekCells = append(currentWeekCells, layoutSpacer(10))
		currentWeekData = append(currentWeekData, models.DailyEntry{})
	}

	// Render day cells
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-02")

		// Check data for this day
		var entry models.DailyEntry

		//Existing data
		if existingSheet != nil && len(existingSheet.Entries) > 0 {
			if val, ok := existingSheet.Entries[dateStr]; ok {
				entry = val
			}
		} else {
			// Autofill with schedule
			dayOfWeek := int(date.Weekday()) - 1
			if dayOfWeek < 0 {dayOfWeek = 6}

			// Get schedule and calculate hours based on time ranges from schedule
			if sched, ok := c.Profile.Schedule[dayOfWeek]; ok && sched.Active {
				//Calculation loop
				var hours float64
				for _, r := range sched.Ranges {
					hours += CalculateDailyHours(r.Start + "-" + r.End)
				}
				entry.HoursWorked = hours
			}
		}

		// ensure date is set
		entry.Date = dateStr

		// Create widget
		cell := NewDayCell(day, entry, c.Profile.Type)
		cell.SetExtrasVisible(c.ShowDetails)
		c.DayWidgets[dateStr] = cell

		//Add individual cells with data
		currentWeekCells = append(currentWeekCells, cell.CanvasObj)
		currentWeekData = append(currentWeekData, entry)

		// End of week check
		if len(currentWeekCells) == 7 {
			c.renderWeekRow(currentWeekCells, currentWeekData)
			currentWeekCells = nil
			currentWeekData = nil
		}
	}

	//Handle End of month padding
	if len(currentWeekCells) > 0 {
		for len(currentWeekCells) < 7 {
			currentWeekCells = append(currentWeekCells, layoutSpacer(10))
			currentWeekData = append(currentWeekData, models.DailyEntry{})
		}

		c.renderWeekRow(currentWeekCells, currentWeekData)
	}

	c.WeeksContainer.Refresh()

	// Show monthly total
	c.updateFooterTotals(existingSheet)
}

//TODO: Row rendering helper
func (c *CalendarPage) renderWeekRow(cells []fyne.CanvasObject, data []models.DailyEntry) {
	//Build grid
	dayGrid := container.NewGridWithColumns(7)
	for _, obj := range cells {dayGrid.Add(obj)}

	//Calculate weekly totals
	statsBox := c.buildWeeklyStats(data)

	//Combine weekly entries with totals
	split := container.NewHSplit(dayGrid, statsBox)
	split.SetOffset(0.85)

	//Add week container
	c.WeeksContainer.Add(split)
	c.WeeksContainer.Add(widget.NewSeparator())
}

//TODO:: build weekly stats helper
func (c *CalendarPage) buildWeeklyStats(data []models.DailyEntry) fyne.CanvasObject {
	//Add each total
	var totalWorked, totalSick, totalVac, totalHol, totalComp, totalOther float64
	for _, d := range data {
		totalWorked += d.HoursWorked
		totalSick += d.SickLeave
		totalVac += d.Vacation
		totalHol += d.Holiday
		totalComp += d.CompTimeTaken // extra hours worked input
		totalOther += d.OtherPaid
	}
	
	//Total weekly compensation
	totalCompensated := totalWorked + totalSick + totalVac + totalHol + totalComp + totalOther
	
	//Get overtime based on threshodls from models/employee.go
	threshold := c.Profile.Type.OvertimeThreshold()
	var otHours float64
	if totalCompensated > threshold {
		otHours = totalCompensated - threshold
	}

	// Build the text UI
	content := container.NewVBox()

	//Line 1: Hours Worked
	content.Add(widget.NewLabelWithStyle(fmt.Sprintf("Worked %.2f", totalWorked), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	if c.Profile.Type == models.TypeFullTime {
		details := ""
		if totalSick > 0 {details += fmt.Sprintf("Sick: %.1f\n", totalSick)}
		if totalVac > 0 {details += fmt.Sprintf("Vac: %.1f\n", totalVac)}
		if totalHol > 0 {details += fmt.Sprintf("Hol: %.f\n", totalHol)}
		if totalComp > 0 {details += fmt.Sprintf("Comp: %.1f\n", totalComp)}
		if totalOther > 0 {details += fmt.Sprintf("Other: %.1f\n")}

		if details != "" {
			label := widget.NewLabel(details)
			label.Wrapping = fyne.TextWrapWord
			content.Add(label)
		}
	}

	// Include overtime rendering
	if otHours > 0 {
		otLabel := widget.NewLabelWithStyle(fmt.Sprintf("OT: %.2f", otHours), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		content.Add(otLabel)
	}else{
		content.Add(widget.NewLabel("OT: 0.00"))
	}

	return container.NewPadded(content)
}

//TODO: Helper to update monthly totals
func (c *CalendarPage) updateFooterTotals(sheet *models.Timesheet) {
	var grandTotalWork float64

	for _, cell := range c.DayWidgets {
		d:= cell.GetData()
		grandTotalWork += d.HoursWorked
	}

	c.FooterLabel.SetText(fmt.Sprintf("Monthly Total Worked: %.2f hrs", grandTotalWork))
}

// Helper to update month ladle
func (c *CalendarPage) updateMonthLabel() {
	c.MonthLabel.SetText(c.CurrentDate.Format("January 2006"))
}

// Save data
func (c *CalendarPage) saveData() {
	//Collect data from GU
	entries := make(map[string]models.DailyEntry)
	var totalWorked float64
	for dateStr, cell := range c.DayWidgets {
		data := cell.GetData()
		entries[dateStr] = data
		totalWorked += data.HoursWorked
	}

	// Create timesheet model
	ts := models.Timesheet{
		Month:       int(c.CurrentDate.Month()),
		Year:        c.CurrentDate.Year(),
		Entries:     entries,
		TotalWorked: totalWorked,
	}

	if err := c.Repo.SaveTimesheet(ts); err != nil {
		dialog.ShowError(err, c.Window)
		return
	}

	dialog.ShowInformation("Saved", "Timesheet Updated Successfully.", c.Window)
}

// Will work on it later
func (c *CalendarPage) exportData() {
	dialog.ShowInformation("Export", "Export functionality not implemented yet.", c.Window)
}
