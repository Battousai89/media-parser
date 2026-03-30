package parser

import (
	"mime"
	"path/filepath"
	"strings"
)

type MediaTypeDetector struct {
	extensionMap map[string]string
	mimeMap      map[string]string
}

func NewMediaTypeDetector() *MediaTypeDetector {
	d := &MediaTypeDetector{
		extensionMap: make(map[string]string),
		mimeMap:      make(map[string]string),
	}
	d.initMaps()
	return d
}

func (d *MediaTypeDetector) initMaps() {
	images := []string{"jpg", "jpeg", "png", "webp", "gif", "svg", "bmp", "ico", "avif", "heic"}
	for _, ext := range images {
		d.extensionMap[ext] = "image"
	}
	d.mimeMap["image/jpeg"] = "image"
	d.mimeMap["image/png"] = "image"
	d.mimeMap["image/gif"] = "image"
	d.mimeMap["image/webp"] = "image"
	d.mimeMap["image/svg+xml"] = "image"
	d.mimeMap["image/bmp"] = "image"
	d.mimeMap["image/x-icon"] = "image"
	d.mimeMap["image/avif"] = "image"
	d.mimeMap["image/heic"] = "image"

	videos := []string{"mp4", "webm", "ogv", "avi", "mov", "mkv", "wmv", "flv", "m4v"}
	for _, ext := range videos {
		d.extensionMap[ext] = "video"
	}
	d.mimeMap["video/mp4"] = "video"
	d.mimeMap["video/webm"] = "video"
	d.mimeMap["video/ogg"] = "video"
	d.mimeMap["video/x-msvideo"] = "video"
	d.mimeMap["video/quicktime"] = "video"
	d.mimeMap["video/x-matroska"] = "video"

	audios := []string{"mp3", "ogg", "wav", "flac", "aac", "m4a", "opus", "weba"}
	for _, ext := range audios {
		d.extensionMap[ext] = "audio"
	}
	d.mimeMap["audio/mpeg"] = "audio"
	d.mimeMap["audio/ogg"] = "audio"
	d.mimeMap["audio/wav"] = "audio"
	d.mimeMap["audio/flac"] = "audio"
	d.mimeMap["audio/mp4"] = "audio"
	d.mimeMap["audio/webm"] = "audio"

	d.extensionMap["pdf"] = "document"
	d.extensionMap["doc"] = "document"
	d.extensionMap["docx"] = "document"
	d.extensionMap["txt"] = "document"
	d.extensionMap["rtf"] = "document"
	d.extensionMap["odt"] = "document"
	d.mimeMap["application/pdf"] = "document"
	d.mimeMap["application/msword"] = "document"
	d.mimeMap["application/vnd.openxmlformats-officedocument.wordprocessingml.document"] = "document"
	d.mimeMap["text/plain"] = "document"
	d.mimeMap["text/rtf"] = "document"

	archives := []string{"zip", "rar", "7z", "tar", "gz", "bz2", "xz"}
	for _, ext := range archives {
		d.extensionMap[ext] = "archive"
	}
	d.mimeMap["application/zip"] = "archive"
	d.mimeMap["application/x-rar-compressed"] = "archive"
	d.mimeMap["application/x-7z-compressed"] = "archive"
	d.mimeMap["application/x-tar"] = "archive"
	d.mimeMap["application/gzip"] = "archive"
}

func (d *MediaTypeDetector) DetectByURL(url string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(url), "."))
	if mediaType, ok := d.extensionMap[ext]; ok {
		return mediaType
	}
	return "other"
}

func (d *MediaTypeDetector) DetectByMIME(mimeType string) string {
	mainType, _, _ := mime.ParseMediaType(mimeType)
	mainType = strings.ToLower(mainType)

	if mediaType, ok := d.mimeMap[mainType]; ok {
		return mediaType
	}

	parts := strings.Split(mainType, "/")
	if len(parts) == 2 {
		switch parts[0] {
		case "image", "video", "audio":
			return parts[0]
		}
	}

	return "other"
}

func (d *MediaTypeDetector) Detect(url, mimeType string) string {
	if mimeType != "" {
		if t := d.DetectByMIME(mimeType); t != "other" {
			return t
		}
	}

	return d.DetectByURL(url)
}

func (d *MediaTypeDetector) GetExtension(url string) string {
	ext := filepath.Ext(url)
	return strings.TrimPrefix(strings.ToLower(ext), ".")
}

func (d *MediaTypeDetector) IsValidExtension(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	_, ok := d.extensionMap[ext]
	return ok
}
