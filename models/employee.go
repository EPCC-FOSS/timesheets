package models

type EmployeeType string

const (
	TypeFullTime EmployeeType = "Full-Time"
	TypePartTime EmployeeType = "Part-Time"
	TypeWorkStudy EmployeeType = "Work-Study"
)

func (e EmployeeType) OvertimeThreshold() float64 {
	switch e{
	case TypeWorkStudy:
		return 15.0
	case TypePartTime:
		return 19.0
	default:
		return 40.0
	}
}