package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// VideoInfo содержит информацию о видео/аудио с YouTube
type VideoInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Artist      string       `json:"artist"`
	Channel     string       `json:"channel"`
	Duration    int          `json:"duration"` // секунды
	Thumbnail   string       `json:"thumbnail"`
	Description string       `json:"description"`
	Formats     []FormatInfo `json:"formats"`
	AudioURL    string       `json:"audio_url"` // лучший audio-only URL
	VideoURL    string       `json:"video_url"` // лучший video URL (опционально)
}

// FormatInfo содержит информацию о формате потока
type FormatInfo struct {
	FormatID  string  `json:"format_id"`
	Ext       string  `json:"ext"`
	URL       string  `json:"url"`
	ACodec    string  `json:"acodec"`
	VCodec    string  `json:"vcodec"`
	ABR       float64 `json:"abr"` // audio bitrate (kbps)
	VBR       float64 `json:"vbr"` // video bitrate (kbps)
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Protocol  string  `json:"protocol"`
	Format    string  `json:"format"`
}

// Parser для работы с YouTube через yt-dlp
type Parser struct {
	timeout time.Duration
}

// NewParser создаёт новый экземпляр парсера
func NewParser(timeout time.Duration) *Parser {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Parser{
		timeout: timeout,
	}
}

// IsYouTubeURL проверяет, является ли URL ссылкой на YouTube
func IsYouTubeURL(url string) bool {
	youtubePatterns := []string{
		`^(https?://)?(www\.)?(youtube\.com|music\.youtube\.com)/`,
		`^(https?://)?(www\.)?youtu\.be/`,
		`^(https?://)?(music\.)?youtube\.com/`,
	}

	for _, pattern := range youtubePatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return true
		}
	}
	return false
}

// ExtractVideoID извлекает ID видео из YouTube URL
func ExtractVideoID(url string) string {
	patterns := []struct {
		regex   string
		groupID int
	}{
		{`(?:youtube\.com|music\.youtube\.com)/watch\?v=([a-zA-Z0-9_-]{11})`, 1},
		{`youtu\.be/([a-zA-Z0-9_-]{11})`, 1},
		{`youtube\.com/embed/([a-zA-Z0-9_-]{11})`, 1},
		{`youtube\.com/v/([a-zA-Z0-9_-]{11})`, 1},
		{`youtube\.com/shorts/([a-zA-Z0-9_-]{11})`, 1},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(url)
		if len(matches) > p.groupID {
			return matches[p.groupID]
		}
	}

	return ""
}

// GetVideoInfo получает информацию о видео через yt-dlp
func (p *Parser) GetVideoInfo(ctx context.Context, url string) (*VideoInfo, error) {
	videoID := ExtractVideoID(url)
	if videoID == "" {
		return nil, fmt.Errorf("invalid YouTube URL: cannot extract video ID")
	}

	// yt-dlp аргументы для получения JSON с информацией
	args := []string{
		"--dump-json",
		"--no-download",
		"--no-warnings",
		"--format", "bestaudio/best",
		url,
	}

	logArgs := make([]string, len(args))
	copy(logArgs, args)
	// Скрываем URL из логов для безопасности
	for i, arg := range logArgs {
		if arg == url {
			logArgs[i] = "<url>"
		}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Запускаем yt-dlp через sh для корректной работы в Alpine
	// Экранируем URL для shell
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, "&") || strings.Contains(arg, "?") || strings.Contains(arg, "=") {
			escapedArgs[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
		} else {
			escapedArgs[i] = arg
		}
	}
	commandStr := "yt-dlp " + strings.Join(escapedArgs, " ")
	log.Printf("yt-dlp command: %s", commandStr)
	cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", commandStr)
	cmd.Env = append(os.Environ(), "HOME=/tmp", "PYTHONUNBUFFERED=1")
	cmd.Dir = "/tmp"
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := stderrBuf.String()
			log.Printf("yt-dlp stderr: %s", stderr)
			return nil, fmt.Errorf("yt-dlp exited with code %d: %s", exitErr.ExitCode(), stderr)
		}
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yt-dlp timeout after %v", p.timeout)
		}
		return nil, fmt.Errorf("execute yt-dlp: %w", err)
	}

	output := stdoutBuf.Bytes()
	if len(output) == 0 {
		return nil, fmt.Errorf("yt-dlp returned empty output")
	}

	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("parse yt-dlp output: %w. Output: %s", err, string(output)[:min(len(output), 500)])
	}

	// Извлекаем лучший audio URL
	info.AudioURL = p.extractBestAudioURL(&info)

	// Пытаемся извлечь артиста из metadata
	info.Artist = p.extractArtist(&info)

	return &info, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetVideoInfoWithFormats получает полную информацию со всеми форматами
func (p *Parser) GetVideoInfoWithFormats(ctx context.Context, url string) (*VideoInfo, error) {
	videoID := ExtractVideoID(url)
	if videoID == "" {
		return nil, fmt.Errorf("invalid YouTube URL: cannot extract video ID")
	}

	args := []string{
		"--dump-json",
		"--no-download",
		"--no-warnings",
		"--format", "bestaudio+bestvideo/best",
		url,
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", "yt-dlp "+strings.Join(args, " "))
	cmd.Env = append(os.Environ(), "HOME=/tmp", "PYTHONUNBUFFERED=1")
	cmd.Dir = "/tmp"
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	log.Printf("yt-dlp command: yt-dlp %s", strings.Join(args, " "))
	err := cmd.Run()
	log.Printf("yt-dlp stdout length: %d, stderr length: %d", stdoutBuf.Len(), stderrBuf.Len())
	if err != nil {
		stderr := stderrBuf.String()
		log.Printf("yt-dlp stderr: %s", stderr)
		if _, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp error: %s", stderr)
		}
		if cmdCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yt-dlp timeout after %v", p.timeout)
		}
		return nil, fmt.Errorf("execute yt-dlp: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(stdoutBuf.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("parse yt-dlp output: %w", err)
	}

	info.AudioURL = p.extractBestAudioURL(&info)
	info.Artist = p.extractArtist(&info)

	return &info, nil
}

// extractBestAudioURL извлекает лучший audio-only URL из форматов
func (p *Parser) extractBestAudioURL(info *VideoInfo) string {
	if len(info.Formats) == 0 {
		return ""
	}

	var bestAudio FormatInfo
	bestBitrate := 0.0

	for _, format := range info.Formats {
		// Ищем audio-only форматы (нет видео кодека или он "none")
		isAudioOnly := format.VCodec == "none" ||
			format.VCodec == "" ||
			(format.ACodec != "none" && format.ACodec != "")

		if !isAudioOnly {
			continue
		}

		// Предпочитаем m4a (AAC) как наиболее совместимый
		if format.Ext == "m4a" && format.ABR > bestBitrate {
			bestAudio = format
			bestBitrate = format.ABR
		} else if format.ABR > bestBitrate && bestAudio.FormatID == "" {
			bestAudio = format
			bestBitrate = format.ABR
		}
	}

	// Если не нашли audio-only, пробуем найти любой формат с аудио
	if bestAudio.FormatID == "" {
		for _, format := range info.Formats {
			if format.ACodec != "none" && format.ACodec != "" {
				if bestAudio.FormatID == "" || format.ABR > bestAudio.ABR {
					bestAudio = format
				}
			}
		}
	}

	if bestAudio.FormatID != "" {
		return bestAudio.URL
	}

	return ""
}

// extractArtist пытается извлечь имя артиста из metadata
func (p *Parser) extractArtist(info *VideoInfo) string {
	// Проверяем стандартные поля
	if info.Artist != "" {
		return info.Artist
	}

	if info.Channel != "" {
		return info.Channel
	}

	// Пытаемся извлечь из названия (формат "Artist - Title")
	if idx := strings.Index(info.Title, " - "); idx > 0 {
		return strings.TrimSpace(info.Title[:idx])
	}

	return ""
}

// GetDirectAudioURL получает прямой URL на аудио без полной информации
func (p *Parser) GetDirectAudioURL(ctx context.Context, url string) (string, error) {
	args := []string{
		"--get-url",
		"--no-download",
		"--no-warnings",
		"--format", "bestaudio[ext=m4a]/bestaudio/best",
		url,
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", "yt-dlp "+strings.Join(args, " "))
	cmd.Env = append(os.Environ(), "HOME=/tmp", "PYTHONUNBUFFERED=1")
	cmd.Dir = "/tmp"
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	log.Printf("yt-dlp command: yt-dlp %s", strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		stderr := stderrBuf.String()
		log.Printf("yt-dlp stderr: %s", stderr)
		if _, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("yt-dlp error: %s", stderr)
		}
		if cmdCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("yt-dlp timeout after %v", p.timeout)
		}
		return "", fmt.Errorf("execute yt-dlp: %w", err)
	}

	audioURL := strings.TrimSpace(stdoutBuf.String())
	if audioURL == "" {
		return "", fmt.Errorf("no audio URL returned by yt-dlp")
	}

	return audioURL, nil
}

// DownloadAudio скачивает аудио через yt-dlp + ffmpeg в указанный каталог
// Возвращает путь к файлу и размер в байтах
func (p *Parser) DownloadAudio(ctx context.Context, url string, outputDir string) (string, int64, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", 0, fmt.Errorf("create output dir: %w", err)
	}

	// Шаблон имени файла: ID_Название.формат
	outputTemplate := filepath.Join(outputDir, "%(id)s_%(title)s.%(ext)s")

	args := []string{
		"--no-warnings",
		"--extract-audio",
		"--audio-format", "mp3",
		"--output", outputTemplate,
		"--no-playlist",
		url,
	}

	cmdCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", "yt-dlp "+strings.Join(args, " "))
	cmd.Env = append(os.Environ(), "HOME=/tmp", "PYTHONUNBUFFERED=1")
	cmd.Dir = "/tmp"
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	log.Printf("yt-dlp command: yt-dlp %s", strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		stderr := stderrBuf.String()
		log.Printf("yt-dlp stderr: %s", stderr)
		if _, ok := err.(*exec.ExitError); ok {
			return "", 0, fmt.Errorf("yt-dlp error: %s", stderr)
		}
		if cmdCtx.Err() == context.DeadlineExceeded {
			return "", 0, fmt.Errorf("yt-dlp timeout after %v", p.timeout)
		}
		return "", 0, fmt.Errorf("execute yt-dlp: %w", err)
	}

	// yt-dlp выводит имя сохранённого файла в stdout
	filePath := strings.TrimSpace(stdoutBuf.String())
	if filePath == "" {
		return "", 0, fmt.Errorf("yt-dlp did not return file path")
	}

	// Получаем размер файла
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("get file info: %w", err)
	}

	return filePath, fileInfo.Size(), nil
}

// GetFreshAudioURL генерирует свежий URL для скачивания (URL YouTube живут ~6 часов)
func (p *Parser) GetFreshAudioURL(ctx context.Context, videoID string) (string, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	return p.GetDirectAudioURL(ctx, url)
}

// ExtractVideoIDFromURL извлекает ID из сохранённого URL YouTube
func ExtractVideoIDFromURL(url string) string {
	return ExtractVideoID(url)
}

// IsInstalled проверяет, установлен ли yt-dlp в системе
func IsInstalled() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

// GetVersion возвращает версию yt-dlp
func GetVersion() (string, error) {
	cmd := exec.Command("/bin/sh", "-c", "yt-dlp --version")
	cmd.Env = append(os.Environ(), "HOME=/tmp", "PYTHONUNBUFFERED=1")
	cmd.Dir = "/tmp"
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
