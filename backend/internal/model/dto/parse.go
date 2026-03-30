package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type ParseURLRequest struct {
	URL          string   `json:"url" binding:"required,url"`
	MediaType    *string  `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Limit        int      `json:"limit" binding:"required,min=1,max=1000"`
	Offset       int      `json:"offset" binding:"omitempty,min=0"`
	Priority     int      `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int    `json:"media_type_ids,omitempty"`
}

type ParseBatchRequest struct {
	URLs         []string `json:"urls" binding:"required,min=1,max=100"`
	MediaType    *string  `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Limit        int      `json:"limit" binding:"required,min=1,max=1000"`
	Offset       int      `json:"offset" binding:"omitempty,min=0"`
	Priority     int      `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int    `json:"media_type_ids,omitempty"`
}

type ParseAllRequest struct {
	MediaType    *string `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Limit        int     `json:"limit" binding:"required,min=1,max=1000"`
	Offset       int     `json:"offset" binding:"omitempty,min=0"`
	Priority     int     `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int   `json:"media_type_ids,omitempty"`
}

type ParseFirstRequest struct {
	MediaType    *string `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Priority     int     `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int   `json:"media_type_ids,omitempty"`
}

type ParseNRequest struct {
	Count        int     `json:"count" binding:"required,min=1,max=1000"`
	MediaType    *string `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Offset       int     `json:"offset" binding:"omitempty,min=0"`
	Priority     int     `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int   `json:"media_type_ids,omitempty"`
}

type ParseSourceRequest struct {
	SourceID     int     `json:"source_id" binding:"required,min=1"`
	MediaType    *string `json:"media_type,omitempty" binding:"omitempty,oneof=image video audio document archive other"`
	Limit        int     `json:"limit" binding:"required,min=1,max=1000"`
	Offset       int     `json:"offset" binding:"omitempty,min=0"`
	Priority     int     `json:"priority" binding:"omitempty,min=0,max=10"`
	MediaTypeIDs []int   `json:"media_type_ids,omitempty"`
}

type ParseResponse struct {
	RequestID uuid.UUID `json:"request_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

func (r *ParseBatchRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}

func (r *ParseURLRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}

func (r *ParseAllRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}

func (r *ParseFirstRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}

func (r *ParseNRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}

func (r *ParseSourceRequest) MediaTypeIDsToStrings(ids []int) []string {
	codes := make([]string, 0, len(ids))
	for _, id := range ids {
		codes = append(codes, fmt.Sprintf("id:%d", id))
	}
	return codes
}
