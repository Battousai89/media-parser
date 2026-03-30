package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

type Config struct {
	Headless        bool
	PageLoadTimeout time.Duration
	WindowWidth     int
	WindowHeight    int
}

type Driver struct {
	config Config
}

func NewDriver(cfg Config) *Driver {
	if cfg.PageLoadTimeout == 0 {
		cfg.PageLoadTimeout = 30 * time.Second
	}
	if cfg.WindowWidth == 0 {
		cfg.WindowWidth = 1920
	}
	if cfg.WindowHeight == 0 {
		cfg.WindowHeight = 1080
	}
	return &Driver{
		config: cfg,
	}
}

func (d *Driver) GetPageSource(ctx context.Context, url string) (string, error) {
	opts := d.getAllocatorOptions()
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Добавляем таймаут на навигацию
	taskCtx, cancel = context.WithTimeout(taskCtx, d.config.PageLoadTimeout)
	defer cancel()

	var html string
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return "", fmt.Errorf("get page source: %w", err)
	}

	return html, nil
}

func (d *Driver) getAllocatorOptions() []chromedp.ExecAllocatorOption {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Headless,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.WindowSize(d.config.WindowWidth, d.config.WindowHeight),
	}

	if d.config.Headless {
		opts = append(opts, chromedp.Flag("headless", true))
	}

	return opts
}
