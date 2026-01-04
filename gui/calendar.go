package gui

import (
	"log"
	"fmt"
	"image/color"
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
	MonthLabel     *widget.Label
	WeeksContainer *fyne.Container
	FooterLabel    *widget.Label
	ToggleBtn      *widget.Button

	// Data Management
	DayWidgets        map[string]*DayCell
	WeeklyStatsLabels []*widget.Label
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
	c.FooterLabel = widget.NewLabel("Loading...")

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

	return container.NewBorder(
		mainHeader,
		nil, nil, nil,
		container.NewBorder(
			c.buildWeekHeader(),
			c.FooterLabel,
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

	// Build stats header and firce it to specific width
	statsLabel := widget.NewLabelWithStyle("Weekly Stats", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	rightContainer := c.makeFixedContainer(statsLabel)
	
	return container.NewBorder(nil, nil, nil, rightContainer, dayGrid)
}

func (c *CalendarPage) Refresh() {
	c.updateMonthLabel()
	c.WeeksContainer.Objects = nil
	c.DayWidgets = make(map[string]*DayCell)
	c.WeeklyStatsLabels = nil

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
	if startOffset < 0 { startOffset = 6 }

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
			if dayOfWeek < 0 { dayOfWeek = 6 }
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
	// OPTIMIZATION: Set footer immediately using loop calculation, skipping recalculateLive() call
	c.FooterLabel.SetText(fmt.Sprintf("Monthly Total Worked: %.2f hrs", grandTotalWork))
}

// renderWeekRow builds the row and calculates initial stats immediately
func (c *CalendarPage) renderWeekRow(cells []fyne.CanvasObject, data []models.DailyEntry) {
	dayGrid := container.NewGridWithColumns(7)
	for _, obj := range cells {
		dayGrid.Add(obj)
	}

	// Calculate stats immediately
	statsText := generateStatsText(data, c.Profile.Type)
	statsLabel := widget.NewLabel(statsText)
	statsLabel.Wrapping = fyne.TextWrapWord
	statsLabel.Alignment = fyne.TextAlignLeading

	c.WeeklyStatsLabels = append(c.WeeklyStatsLabels, statsLabel)

	rightContainer := c.makeFixedContainer(statsLabel)
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
	if startOffset < 0 { startOffset = 6 }

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
			if weekIndex < len(c.WeeklyStatsLabels) {
				c.WeeklyStatsLabels[weekIndex].SetText(generateStatsText(currentWeekData, c.Profile.Type))
			}
			currentWeekData = nil
			weekIndex++
		}
	}

	// Final partial week
	if len(currentWeekData) > 0 && weekIndex < len(c.WeeklyStatsLabels) {
		c.WeeklyStatsLabels[weekIndex].SetText(generateStatsText(currentWeekData, c.Profile.Type))
	}

	c.FooterLabel.SetText(fmt.Sprintf("Monthly Total Worked: %.2f hrs", grandTotalWork))
}

// Shared logic for stats string generation
func generateStatsText(data []models.DailyEntry, empType models.EmployeeType) string {
	var tWork, tSick, tVac, tHol, tComp, tOther float64
	for _, d := range data {
		tWork += d.HoursWorked
		tSick += d.SickLeave
		tVac += d.Vacation
		tHol += d.Holiday
		tComp += d.CompTimeTaken
		tOther += d.OtherPaid
	}

	// FIX: Correct Overtime Math (Was adding Other twice, missing Comp)
	totalCompensated := tWork + tSick + tVac + tHol + tComp + tOther
	
	threshold := empType.OvertimeThreshold()
	var otHours float64
	if totalCompensated > threshold {
		otHours = totalCompensated - threshold
	}

	text := fmt.Sprintf("Work %.2f", tWork)

	if empType == models.TypeFullTime {
		if tSick > 0 { text += fmt.Sprintf("\nSick %.1f", tSick) }
		if tVac > 0 { text += fmt.Sprintf("\nVac: %.1f", tVac) }
		if tHol > 0 { text += fmt.Sprintf("\nHol: %.1f", tHol) }
		if tComp > 0 { text += fmt.Sprintf("\nComp: %.1f", tComp) }
		if tOther > 0 { text += fmt.Sprintf("\nOth: %.1f", tOther) }
	}

	if otHours > 0 {
		text += fmt.Sprintf("\nOT: %.2f", otHours)
	} else {
		text += "\nOT: 0.00"
	}
	return text
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

func (c *CalendarPage)makeFixedContainer(obj fyne.CanvasObject) fyne.CanvasObject {
	// Create transparent spacer
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(StatsColumnWidth, 10))

	// Use stack, spacer pushes width
	return container.NewStack(spacer, container.NewPadded(obj))
}

func (c *CalendarPage) exportData() {
	dialog.ShowInformation("Export", "Export functionality not implemented yet.", c.Window)
}
