package gui

import (
	"encoding/json"
	"fmt"
	"strings"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/models"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProfileExport combines profile and timesheets for export/import
type ProfileExport struct {
	Profile    models.Profile     `json:"profile"`
	Timesheets []models.Timesheet `json:"timesheets"`
}

type ProfilePage struct {
	Repo   *db.Repository
	Window fyne.Window

	// Form widgets
	FirstName   *widget.Entry
	LastName    *widget.Entry
	Middle      *widget.Entry
	EmpID       *widget.Entry
	PositionNum *widget.Entry
	Dept        *widget.Entry
	Title       *widget.Entry
	Rate        *widget.Entry
	Location    *widget.Entry

	// Supervisor and contact information
	SupervisorName  *widget.Entry
	SupervisorPhone *widget.Entry
	EmployeePhone   *widget.Entry
	OfficePhone     *widget.Entry

	//Dynamic fields
	ExtraGroup *fyne.Container
	Fund       *widget.Entry
	Org        *widget.Entry
	Acct       *widget.Entry
	Prog       *widget.Entry

	// Secondary accounting fields (for part-time employees)
	SecondaryGroup *fyne.Container
	Fund2          *widget.Entry
	Org2           *widget.Entry
	Acct2          *widget.Entry
	Prog2          *widget.Entry
	Rate2          *widget.Entry

	//Type selector
	TypeSelect *widget.Select

	//Schedule inputs
	ScheduleInputs map[int]*widget.Entry

	//State buttons
	SaveButton   *widget.Button
	EditButton   *widget.Button
	ExportButton *widget.Button
	ImportButton *widget.Button

	//Locking logic
	IsLocked bool

	// Update field funtion for ./calendar.go Refresh()
	OnSaved func()
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
	p.PositionNum = widget.NewEntry()
	p.PositionNum.SetPlaceHolder("12345")
	p.Dept = widget.NewEntry()
	p.Dept.SetPlaceHolder("Student Success")
	p.Title = widget.NewEntry()
	p.Title.SetPlaceHolder("Tutor")
	p.Rate = widget.NewEntry()
	p.Rate.SetPlaceHolder("9.67")
	p.Location = widget.NewEntry()
	p.Location.SetPlaceHolder("Main Campus")

	// Supervisor and contact fields
	p.SupervisorName = widget.NewEntry()
	p.SupervisorName.SetPlaceHolder("Jane Smith")
	p.SupervisorPhone = widget.NewEntry()
	p.SupervisorPhone.SetPlaceHolder("(555) 123-4567")
	p.EmployeePhone = widget.NewEntry()
	p.EmployeePhone.SetPlaceHolder("(555) 987-6543")
	p.OfficePhone = widget.NewEntry()
	p.OfficePhone.SetPlaceHolder("(555) 111-2222")

	//Extra fields
	p.Fund = widget.NewEntry()
	p.Org = widget.NewEntry()
	p.Acct = widget.NewEntry()
	p.Prog = widget.NewEntry()

	// Secondary accounting fields
	p.Fund2 = widget.NewEntry()
	p.Org2 = widget.NewEntry()
	p.Acct2 = widget.NewEntry()
	p.Prog2 = widget.NewEntry()
	p.Rate2 = widget.NewEntry()
	p.Rate2.SetPlaceHolder("Optional")

	// Ensure ExtraGroup is initialized so callbacks can safely Show/Hide it
	p.ExtraGroup = container.NewVBox(
		widget.NewLabel("Primary Accounting"),
		widget.NewForm(
			widget.NewFormItem("Fund", p.Fund),
			widget.NewFormItem("Org", p.Org),
			widget.NewFormItem("Account", p.Acct),
			widget.NewFormItem("Program", p.Prog),
		),
	)
	p.ExtraGroup.Hide()

	// Secondary accounting group for part-time employees
	p.SecondaryGroup = container.NewVBox(
		widget.NewLabel("Secondary Accounting (Optional)"),
		widget.NewForm(
			widget.NewFormItem("Fund", p.Fund2),
			widget.NewFormItem("Org", p.Org2),
			widget.NewFormItem("Account", p.Acct2),
			widget.NewFormItem("Program", p.Prog2),
			widget.NewFormItem("Hourly Rate", p.Rate2),
		),
	)
	p.SecondaryGroup.Hide()

	//Dropdown logic
	p.TypeSelect = widget.NewSelect([]string{
		string(models.TypeFullTime),
		string(models.TypePartTime),
		string(models.TypeWorkStudy),
	}, func(selected string) {
		// Show extra fields for pt and ft
		if selected == string(models.TypeWorkStudy) {
			p.ExtraGroup.Hide()
			p.SecondaryGroup.Hide()
		} else if selected == string(models.TypePartTime) {
			p.ExtraGroup.Show()
			p.SecondaryGroup.Show()
		} else {
			// Full-Time
			p.ExtraGroup.Show()
			p.SecondaryGroup.Hide()
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
	p.ExportButton = widget.NewButtonWithIcon("Export", theme.DownloadIcon(), p.exportProfile)
	p.ImportButton = widget.NewButtonWithIcon("Import", theme.UploadIcon(), p.importProfile)
}

func (p *ProfilePage) BuildUI() fyne.CanvasObject {

	//Personal info form
	personalForm := widget.NewForm(
		widget.NewFormItem("First Name", p.FirstName),
		widget.NewFormItem("Last Name", p.LastName),
		widget.NewFormItem("Middle Initial", p.Middle),
		widget.NewFormItem("Employee ID", p.EmpID),
		widget.NewFormItem("Position Number", p.PositionNum),
	)
	personalCard := widget.NewCard("Personal Information", "", personalForm)

	//Common fields
	jobForm := widget.NewForm(
		widget.NewFormItem("Department", p.Dept),
		widget.NewFormItem("Title", p.Title),
		widget.NewFormItem("Location", p.Location),
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
		p.SecondaryGroup,
	))

	//Schedule form

	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	scheduleForm := widget.NewForm()

	for i, day := range days {
		scheduleForm.Append(day, p.ScheduleInputs[i])
	}

	scheduleCard := widget.NewCard("Standard Schedule", "(e.g. 09:00-17:00, 10:00-15:00)", scheduleForm)

	// Supervisor and contact information
	supervisorContactForm := widget.NewForm(
		widget.NewFormItem("Supervisor Name", p.SupervisorName),
		widget.NewFormItem("Supervisor Phone", p.SupervisorPhone),
		widget.NewFormItem("Employee Phone", p.EmployeePhone),
		widget.NewFormItem("Office/Dept Phone", p.OfficePhone),
	)
	supervisorCard := widget.NewCard("Supervisor & Contact Information", "", supervisorContactForm)

	// Buttons
	mainButtons := container.NewGridWithColumns(2, p.SaveButton, p.EditButton)
	importExportButtons := container.NewGridWithColumns(2, p.ExportButton, p.ImportButton)
	buttonRow := container.NewVBox(mainButtons, importExportButtons)

	// Assembled layout for profile
	content := container.NewVBox(
		personalCard,
		jobCard,
		scheduleCard,
		supervisorCard,
		layoutSpacer(10),
		buttonRow,
	)

	return container.NewScroll(container.NewPadded(content))
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
	p.PositionNum.SetText(profile.PositionNum)
	p.Dept.SetText(profile.Department)
	p.Title.SetText(profile.Title)
	p.Rate.SetText(fmt.Sprintf("%.2f", profile.Rate))
	p.Location.SetText(profile.Location)

	// Supervisor and contact info
	p.SupervisorName.SetText(profile.SupervisorName)
	p.SupervisorPhone.SetText(profile.SupervisorPhone)
	p.EmployeePhone.SetText(profile.EmployeePhone)
	p.OfficePhone.SetText(profile.OfficePhone)

	p.TypeSelect.SetSelected(string(profile.Type))

	// Populate primary accounting fields
	p.Fund.SetText(profile.PrimaryAccounting.Fund)
	p.Org.SetText(profile.PrimaryAccounting.Organization)
	p.Acct.SetText(profile.PrimaryAccounting.Account)
	p.Prog.SetText(profile.PrimaryAccounting.Program)

	// Populate secondary accounting fields if they exist
	if profile.SecondaryAccounting != nil {
		p.Fund2.SetText(profile.SecondaryAccounting.Fund)
		p.Org2.SetText(profile.SecondaryAccounting.Organization)
		p.Acct2.SetText(profile.SecondaryAccounting.Account)
		p.Prog2.SetText(profile.SecondaryAccounting.Program)
		if profile.SecondaryAccounting.HourlyRate > 0 {
			p.Rate2.SetText(fmt.Sprintf("%.2f", profile.SecondaryAccounting.HourlyRate))
		}
	}

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

	// Parse secondary accounting rate if provided
	var rate2 float64
	if p.Rate2.Text != "" {
		fmt.Sscanf(p.Rate2.Text, "%f", &rate2)
	}

	// Create profile model
	prof := models.Profile{
		FirstName:       p.FirstName.Text,
		LastName:        p.LastName.Text,
		MiddleInitial:   p.Middle.Text,
		EmployeeID:      p.EmpID.Text,
		PositionNum:     p.PositionNum.Text,
		Department:      p.Dept.Text,
		Title:           p.Title.Text,
		Rate:            rate,
		Location:        p.Location.Text,
		Type:            models.EmployeeType(p.TypeSelect.Selected),
		SupervisorName:  p.SupervisorName.Text,
		SupervisorPhone: p.SupervisorPhone.Text,
		EmployeePhone:   p.EmployeePhone.Text,
		OfficePhone:     p.OfficePhone.Text,
		PrimaryAccounting: models.AccountingCodes{
			Fund:         p.Fund.Text,
			Organization: p.Org.Text,
			Account:      p.Acct.Text,
			Program:      p.Prog.Text,
		},
		Schedule: scheduleMap,
	}

	// Add secondary accounting if any field is filled (for part-time)
	if p.Fund2.Text != "" || p.Org2.Text != "" || p.Acct2.Text != "" || p.Prog2.Text != "" {
		prof.SecondaryAccounting = &models.AccountingCodes{
			Fund:         p.Fund2.Text,
			Organization: p.Org2.Text,
			Account:      p.Acct2.Text,
			Program:      p.Prog2.Text,
			HourlyRate:   rate2,
		}
	}

	// Error saving data
	if err := p.Repo.SaveProfile(&prof); err != nil {
		dialog.ShowError(err, p.Window)
		return
	}

	//Trigger callback for calendar update
	if p.OnSaved != nil {
		p.OnSaved()
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
	p.PositionNum.Disable()
	p.Dept.Disable()
	p.Title.Disable()
	p.Rate.Disable()
	p.Location.Disable()
	p.TypeSelect.Disable()
	p.SupervisorName.Disable()
	p.SupervisorPhone.Disable()
	p.EmployeePhone.Disable()
	p.OfficePhone.Disable()
	p.Fund.Disable()
	p.Org.Disable()
	p.Acct.Disable()
	p.Prog.Disable()
	p.Fund2.Disable()
	p.Org2.Disable()
	p.Acct2.Disable()
	p.Prog2.Disable()
	p.Rate2.Disable()

	for _, entry := range p.ScheduleInputs {
		entry.Disable()
	}
}

func (p *ProfilePage) exportProfile() {
	profile, err := p.Repo.GetProfile()
	if err != nil {
		dialog.ShowError(err, p.Window)
		return
	}
	if profile == nil {
		dialog.ShowInformation("No Profile", "No profile data to export", p.Window)
		return
	}

	// Get all timesheets
	timesheets, err := p.Repo.GetTimesheets()
	if err != nil {
		dialog.ShowError(err, p.Window)
		return
	}

	// Create combined export structure
	exportData := ProfileExport{
		Profile:    *profile,
		Timesheets: timesheets,
	}

	// Create save dialog with default filename and JSON filter
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, p.Window)
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()

		// Marshal combined data to JSON with indentation for readability
		data, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			dialog.ShowError(err, p.Window)
			return
		}

		// Write to file
		if _, err := writer.Write(data); err != nil {
			dialog.ShowError(err, p.Window)
			return
		}

		dialog.ShowInformation("Success", "Profile and timesheets exported successfully", p.Window)
	}, p.Window)

	// Set default filename and file filter
	saveDialog.SetFileName("profile.json")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	saveDialog.Show()
}

func (p *ProfilePage) importProfile() {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, p.Window)
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		// Read file contents
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Try to unmarshal as ProfileExport first (new format)
		var exportData ProfileExport
		if err := json.Unmarshal(data, &exportData); err == nil && exportData.Profile.EmployeeID != "" {
			// New format with timesheets
			// Save imported profile
			if err := p.Repo.SaveProfile(&exportData.Profile); err != nil {
				dialog.ShowError(err, p.Window)
				return
			}

			// Save imported timesheets
			for _, timesheet := range exportData.Timesheets {
				if err := p.Repo.SaveTimesheet(timesheet); err != nil {
					dialog.ShowError(fmt.Errorf("error importing timesheet: %w", err), p.Window)
					// Continue importing other timesheets
				}
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Profile and %d timesheet(s) imported successfully", len(exportData.Timesheets)), p.Window)
		} else {
			// Try old format (profile only)
			var profile models.Profile
			if err := json.Unmarshal(data, &profile); err != nil {
				dialog.ShowError(fmt.Errorf("invalid profile file: %w", err), p.Window)
				return
			}

			// Save imported profile
			if err := p.Repo.SaveProfile(&profile); err != nil {
				dialog.ShowError(err, p.Window)
				return
			}

			dialog.ShowInformation("Success", "Profile imported successfully", p.Window)
		}

		// Reload data to display
		p.LoadData()

		// Trigger callback for calendar update
		if p.OnSaved != nil {
			p.OnSaved()
		}
	}, p.Window)

	// Set file filter to JSON only
	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	openDialog.Show()
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
	p.PositionNum.Enable()
	p.Dept.Enable()
	p.Title.Enable()
	p.Rate.Enable()
	p.Location.Enable()
	p.TypeSelect.Enable()
	p.SupervisorName.Enable()
	p.SupervisorPhone.Enable()
	p.EmployeePhone.Enable()
	p.OfficePhone.Enable()
	p.Fund.Enable()
	p.Org.Enable()
	p.Acct.Enable()
	p.Prog.Enable()
	p.Fund2.Enable()
	p.Org2.Enable()
	p.Acct2.Enable()
	p.Prog2.Enable()
	p.Rate2.Enable()

	// Schedule fields
	for _, entry := range p.ScheduleInputs {
		entry.Enable()
	}
}
