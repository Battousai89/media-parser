package dto

type DownloadURLRequest struct {
	URL string `json:"url" binding:"required,url"`
}
