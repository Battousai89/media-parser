package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/service"
)

type ParseHandler struct {
	parseService *service.ParseService
}

func NewParseHandler(parseService *service.ParseService) *ParseHandler {
	return &ParseHandler{
		parseService: parseService,
	}
}

// @Summary Парсинг одного URL
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseURLRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/url [post]
func (h *ParseHandler) ParseURL(c *gin.Context) {
	var req dto.ParseURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	result, err := h.parseService.ParseURL(c, req.URL, mediaTypeCodes, req.Limit, req.Offset, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "Parse task queued",
		},
	})
}

// @Summary Пакетный парсинг URL
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseBatchRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/batch [post]
func (h *ParseHandler) ParseBatch(c *gin.Context) {
	var req dto.ParseBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	var tokenID *int
	if id, exists := c.Get("token_id"); exists {
		if tid, ok := id.(int); ok {
			tokenID = &tid
		}
	}

	result, err := h.parseService.ParseBatch(c, req.URLs, mediaTypeCodes, req.Limit, req.Offset, req.Priority, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "Batch parse task queued",
		},
	})
}

// @Summary Парсинг всех активных источников
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseAllRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/all [post]
func (h *ParseHandler) ParseAll(c *gin.Context) {
	var req dto.ParseAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	var tokenID *int
	if id, exists := c.Get("token_id"); exists {
		if tid, ok := id.(int); ok {
			tokenID = &tid
		}
	}

	result, err := h.parseService.ParseAllSources(c, mediaTypeCodes, req.Limit, req.Offset, req.Priority, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "All sources parse task queued",
		},
	})
}

// @Summary Первое медиа с каждого источника
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseFirstRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/first [post]
func (h *ParseHandler) ParseFirst(c *gin.Context) {
	var req dto.ParseFirstRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	var tokenID *int
	if id, exists := c.Get("token_id"); exists {
		if tid, ok := id.(int); ok {
			tokenID = &tid
		}
	}

	result, err := h.parseService.ParseFirst(c, mediaTypeCodes, req.Priority, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "First media parse task queued",
		},
	})
}

// @Summary Парсинг N медиа
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseNRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/n [post]
func (h *ParseHandler) ParseN(c *gin.Context) {
	var req dto.ParseNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	var tokenID *int
	if id, exists := c.Get("token_id"); exists {
		if tid, ok := id.(int); ok {
			tokenID = &tid
		}
	}

	result, err := h.parseService.ParseN(c, req.Count, mediaTypeCodes, req.Offset, req.Priority, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "N media parse task queued",
		},
	})
}

// @Summary Парсинг по ID источника
// @Tags parse
// @Accept json
// @Produce json
// @Param request body dto.ParseSourceRequest true "Запрос на парсинг"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/parse/source [post]
func (h *ParseHandler) ParseSource(c *gin.Context) {
	var req dto.ParseSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	mediaTypeCodes := req.MediaTypeIDsToStrings(req.MediaTypeIDs)
	if req.MediaType != nil {
		mediaTypeCodes = append(mediaTypeCodes, *req.MediaType)
	}

	var tokenID *int
	if id, exists := c.Get("token_id"); exists {
		if tid, ok := id.(int); ok {
			tokenID = &tid
		}
	}

	result, err := h.parseService.ParseSource(c, req.SourceID, mediaTypeCodes, req.Limit, req.Offset, req.Priority, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "PARSE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.ParseResponse{
			RequestID: result.ID,
			Status:    "pending",
			Message:   "Source parse task queued",
		},
	})
}
