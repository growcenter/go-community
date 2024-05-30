package models

type Health struct {
	Type	string		`json:"type"`
	Status	interface{}	`json:"status"`
}

type List struct {
	Type		string		`json:"type"`
	Data		interface{}	`json:"data"`
	TotalRows	int			`json:"totalRows"`
}