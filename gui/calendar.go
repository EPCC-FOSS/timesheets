package gui

import (
	"fmt"
	"time"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

type CalendarPage struct {
	Repo *db.Repository
	Window fyne.Window

	//State
	CurrentDate time.Time
	Profile *models.Profile

	//UI components
	MonthLabel *widget.Label
	GridContainer *fyne.Container

	//Daywidgets
	DayWidgets map[string]*Daycell
}

