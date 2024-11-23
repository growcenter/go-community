package models

var (
	TYPE_DEPARTMENT = "department"
)

type DepartmentsResponse struct {
	Type           string `json:"type"`
	DepartmentCode string `json:"departmentCode"`
	DepartmentName string `json:"departmentName"`
}

type CampusesResponse struct {
	Type       string `json:"type"`
	CampusCode string `json:"campusCode"`
	CampusName string `json:"campusName"`
}
