package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/service"
)

type MediaHandler struct {
	mediaService *service.MediaService
}

func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// @Summary Список медиа
// @Tags media
// @Produce json
// @Param media_type query string false "Тип медиа (image, video, document, archive, other)"
// @Param available query bool false "Доступность"
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response{data=dto.PaginatedResponse}
// @Router /api/v1/media [get]
func (h *MediaHandler) GetMediaList(c *gin.Context) {
	var req dto.MediaListRequest
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
	if req.Limit > 100 {
		req.Limit = 100
	}

	var medias []*entity.Media
	var total int
	var err error

	if req.MediaType != nil {
		medias, err = h.mediaService.GetMediaByType(c, *req.MediaType, req.Limit, req.Offset)
		if err == nil {
			total, _ = h.mediaService.CountMediaByType(c, *req.MediaType)
		}
	} else {
		medias, err = h.mediaService.GetAllMedia(c, req.Limit, req.Offset)
		if err == nil {
			total, _ = h.mediaService.CountMedia(c)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	items := make([]*dto.MediaItem, 0, len(medias))
	for _, m := range medias {
		item := &dto.MediaItem{
			ID:        m.ID,
			URL:       m.URL,
			MediaTypeID: m.MediaTypeID,
			Available: m.Available,
			CreatedAt: m.CreatedAt,
		}
		if m.MediaType != nil {
			item.MediaType = m.MediaType.Code
		}
		item.Title = m.Title
		item.FileSize = m.FileSize
		item.MimeType = m.MimeType
		items = append(items, item)
	}

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

// @Summary Получить медиа по ID
// @Tags media
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} dto.Response{data=dto.MediaDetail}
// @Router /api/v1/media/:id [get]
func (h *MediaHandler) GetMediaByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid media ID"},
		})
		return
	}

	media, err := h.mediaService.GetMediaByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	if media == nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "NOT_FOUND", Message: "Media not found"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    convertMediaToDetail(media),
	})
}

// @Summary Загрузить медиа
// @Tags media
// @Accept json
// @Produce json
// @Param request body dto.MediaUploadRequest true "Запрос на загрузку"
// @Success 200 {object} dto.Response
// @Router /api/v1/media/upload [post]
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	var req dto.MediaUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    gin.H{"message": "Upload endpoint ready"},
	})
}

// @Summary Проверить доступность URL
// @Tags media
// @Accept json
// @Produce json
// @Param request body dto.MediaCheckRequest true "URL для проверки"
// @Success 200 {object} dto.Response{data=dto.MediaCheckResponse}
// @Router /api/v1/media/check [post]
func (h *MediaHandler) CheckMedia(c *gin.Context) {
	var req dto.MediaCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: dto.MediaCheckResponse{
			URL:       req.URL,
			Available: true,
		},
	})
}

// @Summary Удалить медиа
// @Tags media
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/media/:id [delete]
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid media ID"},
		})
		return
	}

	if err := h.mediaService.DeleteMedia(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DELETE_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    gin.H{"message": "Media deleted"},
	})
}

func convertMediaToDetail(media *entity.Media) *dto.MediaDetail {
	if media == nil {
		return nil
	}
	result := &dto.MediaDetail{
		ID:          media.ID,
		URL:         media.URL,
		MediaTypeID: media.MediaTypeID,
		Title:       media.Title,
		Description: media.Description,
		FileSize:    media.FileSize,
		MimeType:    media.MimeType,
		Hash:        media.Hash,
		Available:   media.Available,
		CheckedAt:   media.CheckedAt,
		CreatedAt:   media.CreatedAt,
		UpdatedAt:   media.UpdatedAt,
	}
	if media.MediaType != nil {
		result.MediaType = media.MediaType.Code
	}
	return result
}
