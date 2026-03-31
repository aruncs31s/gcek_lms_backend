package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aruncs/esdc-lms/internal/logger"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// ChromePool manages a pool of Chrome contexts for efficient PDF generation
type ChromePoolImpl struct {
	mu       sync.Mutex
	contexts chan context.Context
	AllocCtx context.Context
	cancel   context.CancelFunc
	size     int
	closed   bool
}

var ChromePool *ChromePoolImpl
var once sync.Once

// InitChrome initializes the Chrome pool (call this once at startup)
func InitChrome(poolSize int) error {
	var err error
	once.Do(func() {
		chromePath := getChromeExecutablePath()
		if chromePath == "" {
			err = fmt.Errorf("chrome not found on system placed inside /var/chrome/*")
			return
		}
		opts := getOpts(chromePath)
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), *opts...)
		ChromePool = &ChromePoolImpl{
			contexts: make(chan context.Context, poolSize),
			AllocCtx: allocCtx,
			cancel:   cancel,
			size:     poolSize,
		}

		// Pre-warm the pool with contexts
		for i := 0; i < poolSize; i++ {
			ctx, _ := chromedp.NewContext(allocCtx)
			ChromePool.contexts <- ctx
		}
		logger.GetLogger().Info("Chrome pool initialized", zap.Int("size", poolSize))
	})
	return err
}

func getOpts(chromePath string) *[]chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "OptimizationGuideModelDownloading,IsolateOrigins,site-per-process"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("single-process", false),
		chromedp.Flag("enable-automation", true),
		chromedp.UserDataDir(getTempUserDataDir()),
		chromedp.ExecPath(chromePath),
	)
	return &opts
}

// getTempUserDataDir creates a unique temporary user data directory for Chrome
func getTempUserDataDir() string {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("chromium-user-data-%d", time.Now().Unix()))
	os.MkdirAll(tempDir, 0755)
	return tempDir
}

// CloseChrome closes the Chrome pool (call this on shutdown)
func CloseChrome() {
	if ChromePool != nil {
		ChromePool.mu.Lock()
		defer ChromePool.mu.Unlock()
		if !ChromePool.closed {
			ChromePool.closed = true
			close(ChromePool.contexts)
			ChromePool.cancel()
		}
	}
}

// acquireContext gets a Chrome context from the pool
func (cp *ChromePoolImpl) acquireContext(timeout time.Duration) (context.Context, error) {
	cp.mu.Lock()
	if cp.closed {
		cp.mu.Unlock()
		return nil, fmt.Errorf("chrome pool is closed")
	}
	cp.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case c := <-cp.contexts:
		if c == nil {
			return nil, fmt.Errorf("chrome pool is closed")
		}
		return c, nil
	case <-ctx.Done():
		logger.GetLogger().Warn("Chrome pool exhausted, creating temporary context")
		tempCtx, _ := chromedp.NewContext(cp.AllocCtx)
		return tempCtx, nil
	}
}

// releaseContext returns a Chrome context to the pool
func (cp *ChromePoolImpl) releaseContext(ctx context.Context) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.closed {
		return
	}

	select {
	case cp.contexts <- ctx:
	default:
		// Pool is full
	}
}

// getChromeExecutablePath finds Chrome/Chromium executable on the system
func getChromeExecutablePath() string {
	possiblePaths := []string{
		"/var/chrome/chrome-linux/chrome",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			logger.GetLogger().Info("Found Chrome executable at:", zap.String("path", path))
			return path
		}
	}
	return ""
}

// PDFOptions defines parameters for PDF generation
type PDFOptions struct {
	Landscape       bool
	PrintBackground bool
	PaperWidth      float64
	PaperHeight     float64
	MarginTop       float64
	MarginBottom    float64
	MarginLeft      float64
	MarginRight     float64
}

// DefaultPDFOptions returns standard A4 portrait options
func DefaultPDFOptions() *PDFOptions {
	return &PDFOptions{
		Landscape:       false,
		PrintBackground: true,
		PaperWidth:      8.27,
		PaperHeight:     11.69,
		MarginTop:       0.4,
		MarginBottom:    0.4,
		MarginLeft:      0.4,
		MarginRight:     0.4,
	}
}

func ConvertHTMLToPDF(
	ctx context.Context,
	html string,
	opts *PDFOptions,
) ([]byte, error) {
	if ChromePool == nil {
		return nil, fmt.Errorf("chrome pool not initialized")
	}

	if opts == nil {
		opts = DefaultPDFOptions()
	}

	// Acquire context from pool
	cdCtx, err := ChromePool.acquireContext(5 * time.Second)
	if err != nil {
		return nil, err
	}
	defer ChromePool.releaseContext(cdCtx)

	// Create a new tab context from the pool context
	tabCtx, cancelTab := chromedp.NewContext(cdCtx)
	defer cancelTab()

	// Handle timeout/cancellation
	runCtx, cancelRun := context.WithTimeout(tabCtx, 60*time.Second)
	defer cancelRun()

	var pdf []byte
	err = chromedp.Run(runCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			printOp := page.PrintToPDF().
				WithPrintBackground(opts.PrintBackground).
				WithLandscape(opts.Landscape).
				WithPaperWidth(opts.PaperWidth).
				WithPaperHeight(opts.PaperHeight).
				WithMarginTop(opts.MarginTop).
				WithMarginBottom(opts.MarginBottom).
				WithMarginLeft(opts.MarginLeft).
				WithMarginRight(opts.MarginRight)

			pdf, _, err = printOp.Do(ctx)
			return err
		}),
	)

	if err != nil {
		logger.Log.Warn("local pdf generation failed", zap.Error(err), zap.Int("html_size_bytes", len(html)))
		return nil, err
	}

	return pdf, nil
}
