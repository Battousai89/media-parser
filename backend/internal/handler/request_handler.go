package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/service"
)

type RequestHandler struct {
	requestService *service.RequestService
}

func NewRequestHandler(requestService *service.RequestService) *RequestHandler {
	return &RequestHandler{
		requestService: requestService,
	}
}

// @Summary Получить статус запроса
// @Tags requests
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} dto.Response{data=dto.RequestDetail}
// @Router /api/v1/requests/:id [get]
func (h *RequestHandler) GetRequestByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid request ID"},
		})
		return
	}

	req, err := h.requestService.GetRequestByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	if req == nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "NOT_FOUND", Message: "Request not found"},
		})
		return
	}

	// Проверка прав доступа - пользователь может видеть только свои запросы
	if req.TokenID != nil {
		if tokenID, exists := c.Get("token_id"); exists {
			if tid, ok := tokenID.(int); ok && tid != *req.TokenID {
				c.JSON(http.StatusForbidden, dto.Response{
					Success: false,
					Error:   &dto.ErrorData{Code: "FORBIDDEN", Message: "Access denied to this request"},
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    h.convertRequestToDetail(c, req),
	})
}

// @Summary Список запросов
// @Tags requests
// @Produce json
// @Param status_id query int false "Status ID"
// @Param media_type query string false "Media type"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response{data=dto.PaginatedResponse}
// @Router /api/v1/requests [get]
func (h *RequestHandler) GetRequestList(c *gin.Context) {
	var req dto.RequestListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	var requests []*entity.Request
	var err error

	if req.StatusID != nil {
		requests, err = h.requestService.GetRequestsByStatus(c, *req.StatusID, req.Limit, req.Offset)
	} else {
		requests, err = h.requestService.GetAllRequests(c, req.Limit, req.Offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	items := make([]*dto.RequestItem, 0, len(requests))
	for _, r := range requests {
		sources, _ := h.requestService.GetRequestSources(c, r.ID)
		parsedCount := 0
		for _, s := range sources {
			parsedCount += s.ParsedCount
		}

		item := &dto.RequestItem{
			ID:           r.ID,
			StatusID:     r.StatusID,
			LimitCount:   r.LimitCount,
			OffsetCount:  r.OffsetCount,
			Priority:     r.Priority,
			ParsedCount:  parsedCount,
			SourcesCount: len(sources),
			CreatedAt:    r.CreatedAt,
			CompletedAt:  r.CompletedAt,
		}
		if r.Status != nil {
			item.Status = r.Status.Code
		}
		if len(r.MediaTypeIDs) > 0 {
			item.MediaTypeIDs = r.MediaTypeIDs
		}
		items = append(items, item)
	}

	// Получаем общее количество запросов для пагинации
	total, _ := h.requestService.Count(c)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.PaginatedResponse{
			Items:  items,
			Total:  total,
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	})
}

// @Summary Медиа запроса
// @Tags requests
// @Produce json
// @Param id path string true "Request ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response{data=dto.PaginatedResponse}
// @Router /api/v1/requests/:id/media [get]
func (h *RequestHandler) GetRequestMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid request ID"},
		})
		return
	}

	req, err := h.requestService.GetRequestByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	if req == nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "NOT_FOUND", Message: "Request not found"},
		})
		return
	}

	// Проверка прав доступа
	if req.TokenID != nil {
		if tokenID, exists := c.Get("token_id"); exists {
			if tid, ok := tokenID.(int); ok && tid != *req.TokenID {
				c.JSON(http.StatusForbidden, dto.Response{
					Success: false,
					Error:   &dto.ErrorData{Code: "FORBIDDEN", Message: "Access denied"},
				})
				return
			}
		}
	}

	// Получаем медиа для этого запроса с источниками
	medias, err := h.requestService.GetRequestMediaWithSources(c, id, 1000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	// Если medias nil, создаём пустой массив
	if medias == nil {
		medias = []*dto.RequestMediaItem{}
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.PaginatedResponse{
			Items:  medias,
			Total:  len(medias),
			Limit:  1000,
			Offset: 0,
		},
	})
}

// @Summary Отменить запрос
// @Tags requests
// @Produce json
// @Param id path string true "Request ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/requests/:id [delete]
func (h *RequestHandler) DeleteRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid request ID"},
		})
		return
	}

	if err := h.requestService.DeleteRequest(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DELETE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    gin.H{"message": "Request cancelled"},
	})
}

func (h *RequestHandler) convertRequestToDetail(ctx context.Context, req *entity.Request) *dto.RequestDetail {
	if req == nil {
		return nil
	}

	// Загружаем источники запроса
	requestSources, _ := h.requestService.GetRequestSources(ctx, req.ID)
	sources := make([]dto.RequestSourceItem, 0, len(requestSources))
	for _, rs := range requestSources {
		sourceItem := dto.RequestSourceItem{
			SourceID:     rs.SourceID,
			SourceName:   "",
			BaseURL:      "",
			StatusID:     rs.StatusID,
			MediaCount:   rs.MediaCount,
			ParsedCount:  rs.ParsedCount,
			RetryCount:   rs.RetryCount,
			MaxRetries:   rs.MaxRetries,
			ErrorMessage: rs.ErrorMessage,
		}
		if rs.Source != nil {
			sourceItem.SourceName = rs.Source.Name
			sourceItem.BaseURL = rs.Source.BaseURL
		}
		if rs.Status != nil {
			sourceItem.Status = rs.Status.Code
		}
		sources = append(sources, sourceItem)
	}

	result := &dto.RequestDetail{
		ID:           req.ID,
		StatusID:     req.StatusID,
		LimitCount:   req.LimitCount,
		OffsetCount:  req.OffsetCount,
		Priority:     req.Priority,
		RetryCount:   req.RetryCount,
		MaxRetries:   req.MaxRetries,
		ErrorMessage: req.ErrorMessage,
		StartedAt:    req.StartedAt,
		CompletedAt:  req.CompletedAt,
		CreatedAt:    req.CreatedAt,
		UpdatedAt:    req.UpdatedAt,
		MediaTypeIDs: req.MediaTypeIDs,
		Sources:      sources,
	}
	if req.Status != nil {
		result.Status = req.Status.Code
	}
	return result
}
