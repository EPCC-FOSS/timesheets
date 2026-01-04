package gui

import (
	"fmt"
	"strconv"
	//"strconv"

	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//DayCell struct

type DayCell struct {
	DateStr   string
	CanvasObj fyne.CanvasObject

	//Input
	WorkedEntry *widget.Entry

	ExtrasContainer *fyne.Container

	// Full time inputs
	SickEntry     *widget.Entry
	VacationEntry *widget.Entry
	HolidayEntry  *widget.Entry
	CompEntry     *widget.Entry
	OtherEntry    *widget.Entry
}

func NewDayCell(dayNum int, data models.DailyEntry, empType models.EmployeeType, onChanged func()) *DayCell {
	//Initialize cell
	cell := &DayCell{
		DateStr: data.Date,
	}

	// Generalized input (hurs worked that date)
	cell.WorkedEntry = makeEntry(data.HoursWorked, onChanged)

	//Label for day number
	dayLabel := widget.NewLabelWithStyle(fmt.Sprintf("%d", dayNum), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var content *fyne.Container

	// Full time inputs (accordion inputs for cleaner layout)
	if empType == models.TypeFullTime {
		cell.SickEntry = makeEntry(data.SickLeave, onChanged)
		cell.VacationEntry = makeEntry(data.Vacation, onChanged)
		cell.HolidayEntry = makeEntry(data.Holiday, onChanged)
		cell.CompEntry = makeEntry(data.CompTimeTaken, onChanged)
		cell.OtherEntry = makeEntry(data.OtherPaid, onChanged)

		//Create extras container
		cell.ExtrasContainer = container.NewVBox(
			inputRow("Sick", cell.SickEntry),
			inputRow("Vac", cell.VacationEntry),
			inputRow("Hol", cell.HolidayEntry),
			inputRow("Comp", cell.CompEntry),
			inputRow("Oth", cell.OtherEntry),
		)
		cell.ExtrasContainer.Hide() //HIDE ft button

		//content put together
		content = container.NewVBox(
			container.NewBorder(
				nil, nil,
				dayLabel,
				nil, nil,
			),
			container.NewBorder(
				nil, nil,
				widget.NewLabel("Work:"),
				nil,
				cell.WorkedEntry,
			),
			cell.ExtrasContainer,
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
func makeEntry(val float64, onChanged func()) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("0.0")
	//Cell not empty
	if val > 0 {
		entry.SetText(fmt.Sprintf("%.1f", val))
	}

	//Hook onChange to Fyne's event listener
	entry.OnChanged = func(s string) {
		if onChanged != nil {
			onChanged()
		}
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

// Helper function to toggle extra fields in full time timesheet
func (d *DayCell) SetExtrasVisible (show bool) {
	if d.ExtrasContainer == nil {
		return
	}
	if show {
		d.ExtrasContainer.Show()
	}else{
		d.ExtrasContainer.Hide()
	}
}

// Helper to convert hours worked from text to float
func parseFloat(str string) float64 {
	if str == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

