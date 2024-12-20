package models

var (
	TYPE_DEPARTMENT = "department"
)

type DepartmentsResponse struct {
	Type           string `json:"type" example:"department"`
	DepartmentCode string `json:"departmentCode" example:"TC"`
	DepartmentName string `json:"departmentName" example:"Take Care Department"`
}

type CampusesResponse struct {
	Type       string `json:"type" example:"campus"`
	CampusCode string `json:"campusCode" example:"BKS"`
	CampusName string `json:"campusName" example:"GROW Community Bekasi"`
}
