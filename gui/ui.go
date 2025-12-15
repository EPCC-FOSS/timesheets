package gui

import (
	"fmt"
	"time"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type TimesheetUI struct {
	Window fyne.Window
	Repo *db.Repository
	ListWidget *widget.List
	Data []models.TimesheetEntry
}

