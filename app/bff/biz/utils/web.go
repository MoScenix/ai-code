package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func ScreenshotViewport(target, savePath string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	nav, err := toNavURLLinux(target)
	if err != nil {
		return err
	}
	dir := filepath.Dir(savePath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s failed: %w", dir, err)
		}
	}
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(nav),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		return fmt.Errorf("chromedp run failed: %w", err)
	}

	if dir := filepath.Dir(savePath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s failed: %w", dir, err)
		}
	}

	if err := os.WriteFile(savePath, buf, 0o644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

func toNavURLLinux(target string) (string, error) {
	if strings.Contains(target, "://") {
		return target, nil
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("abs path failed: %w", err)
	}

	// 如果是目录，自动补 index.html
	if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
		abs = filepath.Join(abs, "index.html")
	}

	return "file://" + abs, nil
}
