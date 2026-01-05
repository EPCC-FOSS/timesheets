package pdfgen
/*
import (
	"fmt"
	"os"
	"time"

	"calendar_utility_node_for_timesheets/models"

	"github.com/johnfercher/maroto/color"
	"github.com/johnfercher/maroto/consts"
	"github.com/johnfercher/maroto/pdf"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/props"
)

// Main entrypoint called by UI
func GenerateTimesheet (p *models.Profile, ts *models.Timesheet, outputPath string) error {
	// Init maroto
	m := pdf.NewMaroto(consts.Portrait, consts.Letter)
	m.setPageMargins(20,20,20)

	// Build header
	buildHeader(m, p, ts)

	// Create layout based on profile
	switch p.Type{
	case models.TypeFullTime:
		buildFullTimeTable(m, ts)
	case models.TypePartTime:
		buildPartTimeTable(m, ts)
	case models.TypeWorkStudy:
		buildWorkStudyTable(m, ts)
	}

	//Footer for signatures
	buildFooter(m)

	//Save
	return m.OutputFileAndClose(outputPath)
}

//Create Header (shared header)
func buildHeader(m pdf.Maroto, p *models.Profile, ts *models.Timesheet) {

}
*/