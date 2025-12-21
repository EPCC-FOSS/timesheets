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

	//UI components
	MonthLabel    *widget.Label
	GridContainer *fyne.Container

	//Daywidgets
	DayWidgets map[string]*Daycell
}

// Ui for calendar page
func NewCalendarPage(win fyne.Window, repo *db.Repository) *CalendarPage {
	return &CalendarPage{
		Repo:        repo,
		Window:      win,
		CurrentDate: time.Now(),
		DayWidgets:  make(map[string]*Daycell),
	}
}

func (c *CalendarPage) BuildUI() fyne.CanvasObject {
	c.MonthLabel = widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
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
	exportBtn := widget.NewButtonWithIcon("Export to PDF", theme.DocumentExportIcon(), c.exportData)

	header := container.NewHBox(prevBtn, c.MonthLabel, nextBtn, layoutSpacer(0), saveBtn, exportBtn)

	// Grid for days
	c.GridContainer = container.NewGridWithColumns(7)

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
	c.DayWidgets = make(map[string]*Daycell)

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

}

func (c *CalendarPage) exportData() {
	dialog.ShowInformation("Export", "Export functionality not implemented yet.", c.Window)
}
