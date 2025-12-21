package gui

import (
	"fmt"
	"strings"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type ProfilePage struct {
	Repo   *db.Repository
	Window fyne.Window

	// Form widgets
	FirstName *widget.Entry
	LastName  *widget.Entry
	Middle    *widget.Entry
	EmpID     *widget.Entry
	Dept      *widget.Entry
	Title     *widget.Entry
	Rate      *widget.Entry

	//Dynamic fields
	ExtraGroup *fyne.Container
	Fund       *widget.Entry
	Org        *widget.Entry
	Acct       *widget.Entry
	Prog       *widget.Entry

	//Type selector
	TypeSelect *widget.Select

	//Schedule inputs
	ScheduleInputs map[int]*widget.Entry

	//State buttons
	SaveButton *widget.Button
	EditButton *widget.Button

	//Locking logic
	IsLocked bool
}

func NewProfilePage(win fyne.Window, repo *db.Repository) *ProfilePage {
	p := &ProfilePage{
		Repo:           repo,
		Window:         win,
		ScheduleInputs: make(map[int]*widget.Entry),
	}
	p.initWidgets()
	return p
}

// Create empty widgets
func (p *ProfilePage) initWidgets() {
	//Common fields
	p.FirstName = widget.NewEntry()
	p.FirstName.SetPlaceHolder("John")
	p.LastName = widget.NewEntry()
	p.LastName.SetPlaceHolder("Doe")
	p.Middle = widget.NewEntry()
	p.EmpID = widget.NewEntry()
	p.EmpID.SetPlaceHolder("888888888")
	p.Dept = widget.NewEntry()
	p.Dept.SetPlaceHolder("Student Success")
	p.Title = widget.NewEntry()
	p.Title.SetPlaceHolder("Tutor")
	p.Rate = widget.NewEntry()
	p.Rate.SetPlaceHolder("9.67")

	//Extra fields
	p.Fund = widget.NewEntry()
	p.Org = widget.NewEntry()
	p.Acct = widget.NewEntry()
	p.Prog = widget.NewEntry()

	// Ensure ExtraGroup is initialized so callbacks can safely Show/Hide it
	p.ExtraGroup = container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Fund", p.Fund),
			widget.NewFormItem("Org", p.Org),
			widget.NewFormItem("Account", p.Acct),
			widget.NewFormItem("Program", p.Prog),
		),
	)
	p.ExtraGroup.Hide()

	//Dropdown logic
	p.TypeSelect = widget.NewSelect([]string{
		string(models.TypeFullTime),
		string(models.TypePartTime),
		string(models.TypeWorkStudy),
	}, func(selected string) {
		// Show extra fields for pt  and ft
		if selected == string(models.TypeWorkStudy) {
			p.ExtraGroup.Hide()
		} else {
			p.ExtraGroup.Show()
		}
	})

	//Schedule inputs initialization
	for i := 0; i < 7; i++ {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("08:00-12:00, 13:00-17:00")
		p.ScheduleInputs[i] = entry
	}

	// Initialize control buttons so lockForm can safely call methods
	// These will be replaced with functional buttons in BuildUI.
	p.SaveButton = widget.NewButtonWithIcon("Save Profile", theme.DocumentSaveIcon(), p.saveData)
	p.SaveButton.Importance = widget.HighImportance
	p.EditButton = widget.NewButtonWithIcon("Edit Profile", theme.DocumentCreateIcon(), p.unlockForm)
	p.EditButton.Disable()
}

func (p *ProfilePage) BuildUI() fyne.CanvasObject {

	//Personal info form
	personalForm := widget.NewForm(
		widget.NewFormItem("First Name", p.FirstName),
		widget.NewFormItem("Last Name", p.LastName),
		widget.NewFormItem("Middle Initial", p.Middle),
		widget.NewFormItem("Employee ID", p.EmpID),
	)
	personalCard := widget.NewCard("Personal Information", "", personalForm)

	//Common fields
	jobForm := widget.NewForm(
		widget.NewFormItem("Department", p.Dept),
		widget.NewFormItem("Title", p.Title),
	)

	rateTypeGrid := container.NewGridWithColumns(2,
		widget.NewForm(widget.NewFormItem("Hourly Rate", p.Rate)),
		widget.NewForm(widget.NewFormItem("Employee Type", p.TypeSelect)),
	)

	jobCard := widget.NewCard("Job Details", "", container.NewVBox(
		jobForm,
		rateTypeGrid,
		widget.NewSeparator(),
		p.ExtraGroup,
	))

	//Schedule form

	days := []string{"MMonday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	scheduleForm := widget.NewForm()

	for i, day := range days {
		scheduleForm.Append(day, p.ScheduleInputs[i])
	}

	scheduleCard := widget.NewCard("Standard Schedule", "(e.g. 09:00-17:00, 10:00-15:00)", scheduleForm)

	// Buttons
	buttonRow := container.NewGridWithColumns(2, p.SaveButton, p.EditButton)

	// Assembled layout for profile
	content := container.NewVBox(
		personalCard,
		jobCard,
		scheduleCard,
		layoutSpacer(10),
		buttonRow,
	)

	return container.NewPadded(container.NewScroll(content))
}

func layoutSpacer(height float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, height))
	return spacer
}

func (p *ProfilePage) LoadData() {
	profile, err := p.Repo.GetProfile()
	if err != nil {
		dialog.ShowError(err, p.Window)
		return
	}
	if profile == nil {
		return // No data
	}

	//Populate fields if data is available
	p.FirstName.SetText(profile.FirstName)
	p.LastName.SetText(profile.LastName)
	p.Middle.SetText(profile.MiddleInitial)
	p.EmpID.SetText(profile.EmployeeID)
	p.Dept.SetText(profile.Department)
	p.Title.SetText(profile.Title)
	p.Rate.SetText(fmt.Sprintf("%.2f", profile.Rate))

	p.TypeSelect.SetSelected(string(profile.Type))

	// Populate extra fields
	p.Fund.SetText(profile.Fund)
	p.Org.SetText(profile.Org)
	p.Acct.SetText(profile.Acct)
	p.Prog.SetText(profile.Prog)

	// Populate schedule
	for dayIdx, schedule := range profile.Schedule {
		if input, ok := p.ScheduleInputs[dayIdx]; ok {
			// COnvert range to string
			var rangeStrs []string
			for _, r := range schedule.Ranges {
				rangeStrs = append(rangeStrs, fmt.Sprintf("%s-%s", r.Start, r.End))
			}

			input.SetText(strings.Join(rangeStrs, ", "))
		}
	}

	p.lockForm()
}

func (p *ProfilePage) saveData() {
	// Parse hourly rate
	var rate float64
	fmt.Sscanf(p.Rate.Text, "%f", &rate)

	// Parse schedule
	scheduleMap := make(map[int]models.DaySchedule)
	for i, input := range p.ScheduleInputs {
		text := strings.TrimSpace((input.Text))
		if text == "" {
			continue
		}

		// Parse string into ranges
		var ranges []models.TimeRange
		parts := strings.Split(text, ",")
		for _, part := range parts {
			times := strings.Split(strings.TrimSpace(part), "-")

			// Validate time range
			if len(times) == 2 {
				// Add to ranges
				ranges = append(ranges, models.TimeRange{
					Start: strings.TrimSpace(times[0]),
					End:   strings.TrimSpace(times[1]),
				})
			}
		}

		// Map to day schedule
		scheduleMap[i] = models.DaySchedule{
			Active: true,
			Ranges: ranges,
		}
	}

	// Create profile model
	prof := models.Profile{
		FirstName:     p.FirstName.Text,
		LastName:      p.LastName.Text,
		MiddleInitial: p.Middle.Text,
		EmployeeID:    p.EmpID.Text,
		Department:    p.Dept.Text,
		Title:         p.Title.Text,
		Rate:          rate,
		Type:          models.EmployeeType(p.TypeSelect.Selected),
		Fund:          p.Fund.Text,
		Org:           p.Org.Text,
		Acct:          p.Acct.Text,
		Prog:          p.Prog.Text,
		Schedule:      scheduleMap,
	}

	// Error saving data
	if err := p.Repo.SaveProfile(&prof); err != nil {
		dialog.ShowError(err, p.Window)
		return
	}

	dialog.ShowInformation("Success", "Profile Saved", p.Window)
	p.lockForm()
}

func (p *ProfilePage) lockForm() {
	//General locking
	p.IsLocked = true

	//Buton states
	p.SaveButton.Disable()
	p.EditButton.Enable()

	//Field states
	p.FirstName.Disable()
	p.LastName.Disable()
	p.Middle.Disable()
	p.EmpID.Disable()
	p.Dept.Disable()
	p.Title.Disable()
	p.Rate.Disable()
	p.TypeSelect.Disable()
	p.Fund.Disable()
	p.Org.Disable()
	p.Acct.Disable()
	p.Prog.Disable()

	for _, entry := range p.ScheduleInputs {
		entry.Disable()
	}
}

func (p *ProfilePage) unlockForm() {
	//General locking
	p.IsLocked = false

	//Button states
	p.SaveButton.Enable()
	p.EditButton.Disable()

	//Field states
	p.FirstName.Enable()
	p.LastName.Enable()
	p.Middle.Enable()
	p.EmpID.Enable()
	p.Dept.Enable()
	p.Title.Enable()
	p.Rate.Enable()
	p.TypeSelect.Enable()
	p.Fund.Enable()
	p.Org.Enable()
	p.Acct.Enable()
	p.Prog.Enable()

	// Schedule fields
	for _, entry := range p.ScheduleInputs {
		entry.Enable()
	}
}
