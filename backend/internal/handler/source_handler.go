package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/service"
)

type SourceHandler struct {
	sourceService  *service.SourceService
	patternService *service.PatternService
	parseService   *service.ParseService
}

func NewSourceHandler(
	sourceService *service.SourceService,
	patternService *service.PatternService,
	parseService *service.ParseService,
) *SourceHandler {
	return &SourceHandler{
		sourceService:  sourceService,
		patternService: patternService,
		parseService:   parseService,
	}
}

// @Summary Список источников
// @Tags sources
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.SourceItem}
// @Router /api/v1/sources [get]
func (h *SourceHandler) GetSources(c *gin.Context) {
	sources, err := h.sourceService.GetAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	items := make([]dto.SourceItem, 0, len(sources))
	for _, s := range sources {
		items = append(items, dto.SourceItem{
			ID:        s.ID,
			Name:      s.Name,
			BaseURL:   s.BaseURL,
			StatusID:  s.StatusID,
			UpdatedAt: s.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    items,
	})
}

// @Summary Добавить источник
// @Tags sources
// @Accept json
// @Produce json
// @Param request body dto.SourceCreateRequest true "Данные источника"
// @Success 201 {object} dto.Response{data=dto.SourceDetail}
// @Router /api/v1/sources [post]
func (h *SourceHandler) CreateSource(c *gin.Context) {
	var req dto.SourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	source, err := h.sourceService.Create(c, req.Name, req.BaseURL, req.StatusID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "CREATE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Data:    convertSourceToDetail(source),
	})
}

// @Summary Обновить источник
// @Tags sources
// @Accept json
// @Produce json
// @Param id path int true "Source ID"
// @Param request body dto.SourceUpdateRequest true "Данные для обновления"
// @Success 200 {object} dto.Response{data=dto.SourceDetail}
// @Router /api/v1/sources/:id [put]
func (h *SourceHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid source ID"},
		})
		return
	}

	var req dto.SourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	source, err := h.sourceService.Update(c, id, req.Name, req.BaseURL, req.StatusID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "UPDATE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    convertSourceToDetail(source),
	})
}

// @Summary Удалить источник
// @Tags sources
// @Produce json
// @Param id path int true "Source ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/sources/:id [delete]
func (h *SourceHandler) DeleteSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid source ID"},
		})
		return
	}

	if err := h.sourceService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DELETE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    gin.H{"message": "Source deleted"},
	})
}

// @Summary Парсинг конкретного источника
// @Tags sources
// @Accept json
// @Produce json
// @Param id path int true "Source ID"
// @Param request body dto.ParseSourceRequest true "Параметры парсинга"
// @Success 200 {object} dto.Response{data=dto.ParseResponse}
// @Router /api/v1/sources/:id/parse [post]
func (h *SourceHandler) ParseSource(c *gin.Context) {
	idStr := c.Param("id")
	sourceID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid source ID"},
		})
		return
	}

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

	result, err := h.parseService.ParseSource(c, sourceID, mediaTypeCodes, req.Limit, req.Offset, req.Priority, tokenID)
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

// @Summary Список паттернов
// @Tags patterns
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.PatternItem}
// @Router /api/v1/patterns [get]
func (h *SourceHandler) GetPatterns(c *gin.Context) {
	patterns, err := h.patternService.GetAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	items := make([]dto.PatternItem, 0, len(patterns))
	for _, p := range patterns {
		items = append(items, dto.PatternItem{
			ID:          p.ID,
			Name:        p.Name,
			Regex:       p.Regex,
			MediaTypeID: p.MediaTypeID,
			Priority:    p.Priority,
			CreatedAt:   p.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    items,
	})
}

// @Summary Добавить паттерн
// @Tags patterns
// @Accept json
// @Produce json
// @Param request body dto.PatternCreateRequest true "Данные паттерна"
// @Success 201 {object} dto.Response{data=dto.PatternItem}
// @Router /api/v1/patterns [post]
func (h *SourceHandler) CreatePattern(c *gin.Context) {
	var req dto.PatternCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	pattern, err := h.patternService.Create(c, req.Name, req.Regex, req.MediaTypeID, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "CREATE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Data: dto.PatternItem{
			ID:          pattern.ID,
			Name:        pattern.Name,
			Regex:       pattern.Regex,
			MediaTypeID: pattern.MediaTypeID,
			Priority:    pattern.Priority,
			CreatedAt:   pattern.CreatedAt,
		},
	})
}

// @Summary Обновить паттерн
// @Tags patterns
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Param request body dto.PatternUpdateRequest true "Данные для обновления"
// @Success 200 {object} dto.Response{data=dto.PatternItem}
// @Router /api/v1/patterns/:id [put]
func (h *SourceHandler) UpdatePattern(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid pattern ID"},
		})
		return
	}

	var req dto.PatternUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	pattern, err := h.patternService.Update(c, id, req.Name, req.Regex, req.MediaTypeID, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "UPDATE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.PatternItem{
			ID:          pattern.ID,
			Name:        pattern.Name,
			Regex:       pattern.Regex,
			MediaTypeID: pattern.MediaTypeID,
			Priority:    pattern.Priority,
			CreatedAt:   pattern.CreatedAt,
		},
	})
}

// @Summary Удалить паттерн
// @Tags patterns
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/patterns/:id [delete]
func (h *SourceHandler) DeletePattern(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid pattern ID"},
		})
		return
	}

	if err := h.patternService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DELETE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    gin.H{"message": "Pattern deleted"},
	})
}

func convertSourceToDetail(source *entity.Source) *dto.SourceDetail {
	if source == nil {
		return nil
	}
	return &dto.SourceDetail{
		ID:        source.ID,
		Name:      source.Name,
		BaseURL:   source.BaseURL,
		StatusID:  source.StatusID,
		UpdatedAt: source.UpdatedAt,
		CreatedAt: source.CreatedAt,
	}
}
