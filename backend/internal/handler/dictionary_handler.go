package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/service"
)

type DictionaryHandler struct {
	dictionaryService *service.DictionaryService
}

func NewDictionaryHandler(dictionaryService *service.DictionaryService) *DictionaryHandler {
	return &DictionaryHandler{
		dictionaryService: dictionaryService,
	}
}

// @Summary Получить все справочники
// @Tags dictionaries
// @Produce json
// @Success 200 {object} dto.Response{data=dto.DictionaryResponse}
// @Router /api/v1/dictionaries [get]
func (h *DictionaryHandler) GetDictionaries(c *gin.Context) {
	mediaTypes, err := h.dictionaryService.GetMediaTypes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	requestStatuses, err := h.dictionaryService.GetRequestStatuses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	sourceStatuses, err := h.dictionaryService.GetSourceStatuses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	response := dto.DictionaryResponse{
		MediaTypes:      convertToDictionaryItems(mediaTypes),
		RequestStatuses: convertToDictionaryItemsRequest(requestStatuses),
		SourceStatuses:  convertToDictionaryItemsSource(sourceStatuses),
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    response,
	})
}

// @Summary Получить типы медиа
// @Tags dictionaries
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.DictionaryItem}
// @Router /api/v1/dictionaries/media-types [get]
func (h *DictionaryHandler) GetMediaTypes(c *gin.Context) {
	mediaTypes, err := h.dictionaryService.GetMediaTypes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    convertToDictionaryItems(mediaTypes),
	})
}

// @Summary Получить статусы запросов
// @Tags dictionaries
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.DictionaryItem}
// @Router /api/v1/dictionaries/request-statuses [get]
func (h *DictionaryHandler) GetRequestStatuses(c *gin.Context) {
	statuses, err := h.dictionaryService.GetRequestStatuses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    convertToDictionaryItemsRequest(statuses),
	})
}

// @Summary Получить статусы источников
// @Tags dictionaries
// @Produce json
// @Success 200 {object} dto.Response{data=[]dto.DictionaryItem}
// @Router /api/v1/dictionaries/source-statuses [get]
func (h *DictionaryHandler) GetSourceStatuses(c *gin.Context) {
	statuses, err := h.dictionaryService.GetSourceStatuses(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FETCH_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    convertToDictionaryItemsSource(statuses),
	})
}

func convertToDictionaryItems(items []*entity.MediaType) []dto.DictionaryItem {
	result := make([]dto.DictionaryItem, len(items))
	for i, item := range items {
		result[i] = dto.DictionaryItem{
			ID:   item.ID,
			Code: item.Code,
			Name: item.Name,
		}
	}
	return result
}

func convertToDictionaryItemsRequest(items []*entity.RequestStatus) []dto.DictionaryItem {
	result := make([]dto.DictionaryItem, len(items))
	for i, item := range items {
		result[i] = dto.DictionaryItem{
			ID:   item.ID,
			Code: item.Code,
			Name: item.Name,
		}
	}
	return result
}

func convertToDictionaryItemsSource(items []*entity.SourceStatus) []dto.DictionaryItem {
	result := make([]dto.DictionaryItem, len(items))
	for i, item := range items {
		result[i] = dto.DictionaryItem{
			ID:   item.ID,
			Code: item.Code,
			Name: item.Name,
		}
	}
	return result
}
