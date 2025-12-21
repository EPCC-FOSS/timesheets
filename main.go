package main

import (
	"log"
	"os"
	"path/filepath"

	"calendar_utility_node_for_timesheets/db"
	"calendar_utility_node_for_timesheets/gui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Calendar Utility Node for Timesheets")

	//DB SETUP
	configDir, _ := os.UserConfigDir()
	appPath := filepath.Join(configDir, "Timesheets")
	os.MkdirAll(appPath, 0755)

	repo, err := db.NewRepository(appPath)
	if err != nil {
		log.Fatal(err)
	}

	//Setup Pages
	profilePage := gui.NewProfilePage(myWindow, repo)

	//Load data on startup
	profilePage.LoadData()

	//Layout for tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Profile", profilePage.BuildUI()),
		container.NewTabItem("Calendar", container.NewVBox()), //placeholder for calendar tab
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
