package models

var TYPE_DETAIL = "detail"

type Health struct {
	Type   string      `json:"type"`
	Status interface{} `json:"status"`
}

type List struct {
	Type      string      `json:"type" example:"collection"`
	Data      interface{} `json:"data"`
	TotalRows int         `json:"totalRows"`
}

type ListWithDetail struct {
	Type      string      `json:"type" example:"collection"`
	Details   interface{} `json:"details"`
	Data      interface{} `json:"data"`
	TotalRows int         `json:"totalRows"`
}

type Pagination struct {
	Type           string      `json:"type" example:"collection"`
	Details        interface{} `json:"details,omitempty"`
	Data           interface{} `json:"data"`
	PaginationInfo interface{} `json:"pagination"`
}

type CursorPagination[T any] struct {
	Type           string      `json:"type" example:"collection"`
	Details        interface{} `json:"details,omitempty"`
	Data           interface{} `json:"data"`
	PaginationInfo interface{} `json:"pagination"`
}

type PaginationInfo struct {
	CurrentPage int         `json:"page"`
	TotalPages  int         `json:"totalPages"`
	TotalData   int         `json:"totalData"`
	Limit       int         `json:"limit"`
	Parameter   interface{} `json:"parameters,omitempty"`
}

type Cursor struct {
	Type           string      `json:"type" example:"collection"`
	Data           interface{} `json:"data"`
	PaginationInfo interface{} `json:"pagination"`
}

type CursorInfo struct {
	PreviousCursor string      `json:"previous"`
	NextCursor     string      `json:"next"`
	TotalData      int         `json:"totalData"`
	Limit          int         `json:"limit"`
	Parameter      interface{} `json:"parameters,omitempty"`
}

type (
	PaginationOutput struct {
		Prev  string
		Next  string
		Total int
	}
	ErrorResponse struct {
		Code     int         `json:"code"`
		Status   string      `json:"status"`
		Message  string      `json:"message"`
		Errors   interface{} `json:"errors,omitempty"`
		Metadata Metadata    `json:"metadata"`
	}
	Response struct {
		Code       int         `json:"code"`
		Status     string      `json:"status"`
		Message    string      `json:"message"`
		Data       interface{} `json:"data,omitempty"`
		Pagination *CursorInfo `json:"pagination,omitempty"`
		Errors     interface{} `json:"errors,omitempty"`
		Metadata   Metadata    `json:"metadata"`
	}
	Metadata struct {
		RequestId string `json:"requestId"`
		Timestamp string `json:"timestamp"`
	}
)
