package utils

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  bool   `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Meta   Pagination  `json:"meta"`
}

type Pagination struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	NextPage    *int `json:"next_page,omitempty"`
	PrevPage    *int `json:"prev_page,omitempty"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}
