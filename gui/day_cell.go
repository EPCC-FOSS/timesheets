package gui

import (
	"fmt"
	"strconv"
	//"strconv"

	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//DayCell struct

type DayCell struct {
	DateStr   string
	CanvasObj fyne.CanvasObject

	//Input
	WorkedEntry *widget.Entry

	// Full time inputs
	SickEntry     *widget.Entry
	VacationEntry *widget.Entry
	HolidayEntry  *widget.Entry
	CompEntry     *widget.Entry
	OtherEntry    *widget.Entry
}

func NewDayCell(dayNum int, data models.DailyEntry, empType models.EmployeeType) *DayCell {
	//Initialize cell
	cell := &DayCell{
		DateStr: data.Date,
	}

	// Generalized input (hurs worked that date)
	cell.WorkedEntry = makeEntry(data.HoursWorked)

	//Label for day number
	dayLabel := widget.NewLabelWithStyle(fmt.Sprintf("%d", dayNum), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var content *fyne.Container

	// Full time inputs (accordion inputs for cleaner layout)
	if empType == models.TypeFullTime {
		cell.SickEntry = makeEntry(data.SickLeave)
		cell.VacationEntry = makeEntry(data.Vacation)
		cell.HolidayEntry = makeEntry(data.Holiday)
		cell.CompEntry = makeEntry(data.CompTimeTaken)
		cell.OtherEntry = makeEntry(data.OtherPaid)
		//Create extras container
		extras := container.NewVBox(
			inputRow("Sick", cell.SickEntry),
			inputRow("Vac", cell.VacationEntry),
			inputRow("Hol", cell.HolidayEntry),
			inputRow("Comp", cell.CompEntry),
			inputRow("Oth", cell.OtherEntry),
		)
		extras.Hide() //HIDE ft button

		// Toggle button for extras
		toggleBtn := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() {
			if extras.Hidden {
				extras.Show()
			} else {
				extras.Hide()
			}
		})

		//content put together
		content = container.NewVBox(
			container.NewBorder(nil, nil, dayLabel, toggleBtn, nil), // Header
			container.NewBorder(nil, nil, widget.NewLabel("Work:"), nil, cell.WorkedEntry),
			extras,
		)
	} else {
		// Part tume and work study layout
		content = container.NewVBox(
			dayLabel,
			container.NewBorder(nil, nil, widget.NewLabel("Hours:"), nil, cell.WorkedEntry),
		)
	}

	card := widget.NewCard("", "", content)
	cell.CanvasObj = card

	return cell
}

// Create row container
func inputRow(label string, entry *widget.Entry) fyne.CanvasObject {
	return container.NewBorder(nil, nil, widget.NewLabel(label), nil, entry)
}

// Create new day entry with data provided for cell
func makeEntry(val float64) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("0.0")
	//Cell not empty
	if val > 0 {
		entry.SetText(fmt.Sprintf("%.1f", val))
	}

	return entry
}

// Get data from UI and parse it into struct
func (day *DayCell) GetData() models.DailyEntry {

	// Base entry
	entry := models.DailyEntry{
		Date:        day.DateStr,
		HoursWorked: parseFloat(day.WorkedEntry.Text),
	}

	// Verify full time data is filled based on sick leave
	if day.SickEntry != nil {
		entry.SickLeave = parseFloat(day.SickEntry.Text)
		entry.Vacation = parseFloat(day.VacationEntry.Text)
		entry.Holiday = parseFloat(day.HolidayEntry.Text)
		entry.CompTimeTaken = parseFloat(day.CompEntry.Text)
		entry.OtherPaid = parseFloat(day.OtherEntry.Text)
	}

	return entry
}

// Helper to convert hours worked from text to float
func parseFloat(str string) float64 {
	if str == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(str, 64)
	return f
}
