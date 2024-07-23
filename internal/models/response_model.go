package models

var TYPE_DETAIL = "detail"

type Health struct {
	Type   string      `json:"type"`
	Status interface{} `json:"status"`
}

type List struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	TotalRows int         `json:"totalRows"`
}

type ListWithDetail struct {
	Type      string      `json:"type"`
	Details   interface{} `json:"details"`
	Data      interface{} `json:"data"`
	TotalRows int         `json:"totalRows"`
}
