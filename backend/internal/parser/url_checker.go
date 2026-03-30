package parser

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type URLChecker struct {
	client  *http.Client
	timeout time.Duration
}

func NewURLChecker(timeout time.Duration) *URLChecker {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &URLChecker{
		client:  client,
		timeout: timeout,
	}
}

func (c *URLChecker) Check(ctx context.Context, url string) (*URLStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MediaParserBot/1.0)")

	resp, err := c.client.Do(req)
	if err != nil {
		return &URLStatus{
			URL:       url,
			Available: false,
			Error:     err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	available := resp.StatusCode >= 200 && resp.StatusCode < 400

	return &URLStatus{
		URL:           url,
		Available:     available,
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		LastModified:  resp.Header.Get("Last-Modified"),
	}, nil
}

func (c *URLChecker) CheckWithFallback(ctx context.Context, url string) (*URLStatus, error) {
	status, err := c.Check(ctx, url)
	if err != nil {
		return nil, err
	}

	if status.StatusCode == 405 || status.StatusCode == 501 {
		return c.checkWithRange(ctx, url)
	}

	return status, nil
}

func (c *URLChecker) checkWithRange(ctx context.Context, url string) (*URLStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create range request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MediaParserBot/1.0)")
	req.Header.Set("Range", "bytes=0-0")

	resp, err := c.client.Do(req)
	if err != nil {
		return &URLStatus{
			URL:       url,
			Available: false,
			Error:     err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	available := resp.StatusCode == 206 || (resp.StatusCode == 200)

	return &URLStatus{
		URL:           url,
		Available:     available,
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: parseContentRange(resp.Header.Get("Content-Range")),
	}, nil
}

func parseContentRange(header string) int64 {
	if header == "" {
		return 0
	}

	var total int64
	fmt.Sscanf(header, "bytes %*d-%*d/%d", &total)
	return total
}

type URLStatus struct {
	URL           string `json:"url"`
	Available     bool   `json:"available"`
	StatusCode    int    `json:"status_code"`
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int64  `json:"content_length,omitempty"`
	LastModified  string `json:"last_modified,omitempty"`
	Error         string `json:"error,omitempty"`
}

func (c *URLChecker) BatchCheck(ctx context.Context, urls []string, maxConcurrent int) []*URLStatus {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	results := make([]*URLStatus, len(urls))
	sem := make(chan struct{}, maxConcurrent)

	for i, url := range urls {
		sem <- struct{}{}
		go func(idx int, u string) {
			defer func() { <-sem }()
			status, _ := c.Check(ctx, u)
			results[idx] = status
		}(i, url)
	}

	for i := 0; i < maxConcurrent; i++ {
		sem <- struct{}{}
	}

	return results
}
