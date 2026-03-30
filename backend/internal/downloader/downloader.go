package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Timeout       time.Duration
	MaxFileSize   int64 // 0 = без ограничений
	DestDir       string
	UserAgent     string
}

type Downloader struct {
	config Config
	client *http.Client
}

func NewDownloader(cfg Config) *Downloader {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 100 * 1024 * 1024 // 100MB default
	}
	if cfg.DestDir == "" {
		cfg.DestDir = "./downloads"
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (compatible; MediaParserBot/1.0)"
	}

	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &Downloader{
		config: cfg,
		client: client,
	}
}

type DownloadResult struct {
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	Duration  time.Duration `json:"duration"`
}

func (d *Downloader) Download(ctx context.Context, url string) (*DownloadResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", d.config.UserAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	if contentLength > d.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes", contentLength)
	}

	if err := os.MkdirAll(d.config.DestDir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	fileName := generateFileName(url, resp.Header.Get("Content-Type"))
	filePath := filepath.Join(d.config.DestDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	if contentLength <= 0 {
		writer = &LimitedWriter{W: file, N: d.config.MaxFileSize}
	}

	written, err := io.Copy(writer, io.LimitReader(resp.Body, d.config.MaxFileSize))
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("copy data: %w", err)
	}

	return &DownloadResult{
		FilePath: filePath,
		FileSize: written,
		MimeType: resp.Header.Get("Content-Type"),
		Duration: time.Since(start),
	}, nil
}

func (d *Downloader) DownloadToPath(ctx context.Context, url, destPath string) (*DownloadResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, io.LimitReader(resp.Body, d.config.MaxFileSize))
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("copy data: %w", err)
	}

	return &DownloadResult{
		FilePath: destPath,
		FileSize: written,
		MimeType: resp.Header.Get("Content-Type"),
		Duration: time.Since(start),
	}, nil
}

func generateFileName(urlStr, contentType string) string {
	ext := getExtension(contentType)
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}

func getExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "audio/mpeg":
		return ".mp3"
	case "application/pdf":
		return ".pdf"
	default:
		return ".bin"
	}
}

type LimitedWriter struct {
	W io.Writer
	N int64
	written int64
}

func (l *LimitedWriter) Write(p []byte) (int, error) {
	if l.written+int64(len(p)) > l.N {
		return 0, fmt.Errorf("write limit exceeded")
	}
	n, err := l.W.Write(p)
	l.written += int64(n)
	return n, err
}

func (d *Downloader) GetFileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

func (d *Downloader) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
