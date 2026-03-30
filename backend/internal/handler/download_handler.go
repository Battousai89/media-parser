package handler

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/service"
)

type DownloadHandler struct {
	downloadService *service.DownloadService
}

func NewDownloadHandler(downloadService *service.DownloadService) *DownloadHandler {
	return &DownloadHandler{
		downloadService: downloadService,
	}
}

// @Summary Скачать медиа по ID
// @Tags download
// @Produce application/octet-stream
// @Param id path string true "Media ID"
// @Success 200 {file} binary
// @Failure 400 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/v1/download/:id [post]
func (h *DownloadHandler) DownloadMedia(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "INVALID_ID", Message: "Invalid media ID"},
		})
		return
	}

	media, err := h.downloadService.GetMediaByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "MEDIA_NOT_FOUND", Message: "Media not found"},
		})
		return
	}

	log.Printf("Download: media %s, URL: %s, type: %d", id, media.URL, media.MediaTypeID)

	if media.StoragePath == nil || *media.StoragePath == "" {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "FILE_NOT_IN_STORAGE", Message: "File not found in storage"},
		})
		return
	}

	log.Printf("Download: file exists in storage %s", *media.StoragePath)

	// Скачиваем файл из Minio и отдаём клиенту
	reader, err := h.downloadService.GetFileReader(c, *media.StoragePath)
	if err != nil {
		log.Printf("Download: failed to read file from minio: %v", err)
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DOWNLOAD_ERROR", Message: err.Error()},
		})
		return
	}

	// Читаем всё содержимое
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Download: failed to read file data: %v", err)
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   &dto.ErrorData{Code: "DOWNLOAD_ERROR", Message: err.Error()},
		})
		return
	}

	log.Printf("Download success: %s, size: %d bytes", *media.StoragePath, len(data))

	// Определяем Content-Type
	contentType := "application/octet-stream"
	if media.MimeType != nil && *media.MimeType != "" {
		contentType = *media.MimeType
	}

	// Формируем имя файла с расширением
	filename := "download"
	if media.Title != nil && *media.Title != "" {
		filename = *media.Title
	} else {
		// Если title нет, берём имя из URL
		if u, err := url.Parse(media.URL); err == nil {
			base := filepath.Base(u.Path)
			if base != "" && base != "/" {
				// Декодируем URL и берём имя файла
				filename, _ = url.PathUnescape(base)
				// Сокращаем если слишком длинное (макс 50 символов)
				if len(filename) > 50 {
					ext := filepath.Ext(filename)
					name := strings.TrimSuffix(filename, ext)
					filename = name[:40] + ext
				}
			}
		}
	}

	// Добавляем расширение если его нет
	if !strings.Contains(filepath.Base(filename), ".") {
		if media.MimeType != nil && *media.MimeType != "" {
			// Получаем расширение из mime типа
			extensions, _ := mime.ExtensionsByType(*media.MimeType)
			if len(extensions) > 0 {
				filename = filename + extensions[0]
			}
		} else if media.StoragePath != nil {
			// Берём расширение из storage_path
			ext := filepath.Ext(*media.StoragePath)
			if ext != "" {
				filename = filename + ext
			}
		}
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, contentType, data)
}

// @Summary Скачать медиа по URL
// @Tags download
// @Accept json
// @Produce application/octet-stream
// @Param request body dto.DownloadURLRequest true "Запрос на скачивание"
// @Success 200 {file} binary
// @Failure 400 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /api/v1/download/url [post]
func (h *DownloadHandler) DownloadByURL(c *gin.Context) {
	c.JSON(http.StatusBadRequest, dto.Response{
		Success: false,
		Error:   &dto.ErrorData{Code: "NOT_IMPLEMENTED", Message: "Download by URL is not implemented. Use download by media ID."},
	})
}
