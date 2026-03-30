package dto

import "github.com/google/uuid"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginationRequest struct {
	Limit  int `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" json:"offset" binding:"omitempty,min=0"`
}

type PaginatedResponse struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type IDResponse struct {
	ID interface{} `json:"id"`
}

type StatusResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
