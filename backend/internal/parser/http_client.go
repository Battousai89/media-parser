package parser

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type HTTPClientConfig struct {
	Timeout         time.Duration
	MaxConnsPerHost int
	UserAgent       string
}

type HTTPClient struct {
	config HTTPClientConfig
	client *http.Client
}

func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxConnsPerHost == 0 {
		cfg.MaxConnsPerHost = 100
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxConnsPerHost:     cfg.MaxConnsPerHost,
			MaxIdleConns:        cfg.MaxConnsPerHost,
			MaxIdleConnsPerHost: cfg.MaxConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	return &HTTPClient{
		config: cfg,
		client: client,
	}
}

func (c *HTTPClient) Fetch(ctx context.Context, url string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return &FetchResult{
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		Body:          string(body),
		Headers:       resp.Header,
	}, nil
}

func (c *HTTPClient) Head(ctx context.Context, url string) (*HeadResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create head request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("head request: %w", err)
	}
	defer resp.Body.Close()

	return &HeadResult{
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		LastModified:  resp.Header.Get("Last-Modified"),
		Available:     resp.StatusCode >= 200 && resp.StatusCode < 400,
	}, nil
}

type FetchResult struct {
	StatusCode    int
	ContentType   string
	ContentLength int64
	Body          string
	Headers       http.Header
}

type HeadResult struct {
	StatusCode    int
	ContentType   string
	ContentLength int64
	LastModified  string
	Available     bool
}
