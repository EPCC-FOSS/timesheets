package gui

import (
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

	//UI components
	MonthLabel    *widget.Label
	GridContainer *fyne.Container

	//Daywidgets
	DayWidgets map[string]*DayCell
}

// Ui for calendar page
func NewCalendarPage(win fyne.Window, repo *db.Repository) *CalendarPage {
	c := &CalendarPage{
		Repo:        repo,
		Window:      win,
		CurrentDate: time.Now(),
		DayWidgets:  make(map[string]*DayCell),
	}

	//FIX: Init widgets immediately so refresh can use them safely
	c.MonthLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.GridContainer = container.NewGridWithColumns(7)
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

	header := container.NewHBox(prevBtn, c.MonthLabel, nextBtn, layoutSpacer(0), saveBtn, exportBtn)

	// Grid for days
	// No longer need c.GridContainer = container.NewGridWithColumns(7)

	return container.NewBorder(
		header,
		nil, nil, nil,
		container.NewVBox(
			c.buildWeekHeader(),
			container.NewScroll(c.GridContainer),
		),
	)
}

func (c *CalendarPage) buildWeekHeader() fyne.CanvasObject {
	// Weekday headers
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	header := container.NewGridWithColumns(7)
	for _, d := range days {
		header.Add(widget.NewLabelWithStyle(d, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	}
	return header
}

func (c *CalendarPage) Refresh() {
	c.updateMonthLabel()
	c.GridContainer.Objects = nil
	c.DayWidgets = make(map[string]*DayCell)

	// get proofile
	prof, err := c.Repo.GetProfile()
	if err != nil {
		dialog.ShowError(err, c.Window)
		return
	}
	c.Profile = prof

	// Check if existing timesheet for month
	existingSheet, _ := c.Repo.GetTimesheetByDate(int(c.CurrentDate.Month()), int(c.CurrentDate.Year()))

	//Calculate days
	year, month, _ := c.CurrentDate.Date()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	
	// Determine padding for first day (eg, 1st on monday, 
	// 1 cell of padding to the left of the first row)
	startOffset := int(firstOfMonth.Weekday()) - 1
	if startOffset < 0{
		startOffset = 6 //Sunday
	}

	//Add padding
	for i:= 0; i < startOffset; i++ {
		c.GridContainer.Add(layoutSpacer(10))
	}

	// Render days in month
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	for day:= 1; day <= daysInMonth; day++{
		date := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)
		dateStr := date.Format("2006-01-01")

		// Check data for this day
		var entry models.DailyEntry
		//Existing data
		if existingSheet != nil && len(existingSheet.Entries) > 0{
			if val, ok := existingSheet.Entries[dateStr]; ok {
				entry = val
			}
		} else {
			// No existing data
			//Convert go weekday to our map
			dayOfWeek := int(date.Weekday())-1
			if dayOfWeek < 0 {
				dayOfWeek = 6
			}

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
		c.DayWidgets[dateStr] = cell
		c.GridContainer.Add(cell.CanvasObj)
	}

	// Refresh the container
	c.GridContainer.Refresh()
}

// Helper to update month ladle
func (c *CalendarPage) updateMonthLabel(){
	c.MonthLabel.SetText(c.CurrentDate.Format("January 2026"))
}

// Save data
func (c *CalendarPage) saveData(){
	//Collect data from GU
	entries := make(map[string]models.DailyEntry)
	var totalWorked float64
	for dateStr, cell := range c.DayWidgets{
		data := cell.GetData()
		entries[dateStr] = data
		totalWorked += data.HoursWorked
	}

	// Create timesheet model
	ts := models.Timesheet{
		Month: int(c.CurrentDate.Month()),
		Year: c.CurrentDate.Year(),
		Entries: entries,
		TotalWorked: totalWorked,
	}

	if err := c.Repo.SaveTimesheet(ts); err != nil{
		dialog.ShowError(err, c.Window)
		return
	}

	dialog.ShowInformation("Saved","Timesheet Updated Successfully.", c.Window)
}

// Will work on it later
func (c *CalendarPage) exportData() {
	dialog.ShowInformation("Export", "Export functionality not implemented yet.", c.Window)
}
