package hlivetest

import (
	"fmt"
	"sync"

	"github.com/mxschmitt/playwright-go"
)

var (
	browserOnce sync.Once
	pwContext   playwright.BrowserContext
)

func NewBrowserPage() playwright.Page {
	browserOnce.Do(func() {
		pw, err := playwright.Run()
		if err != nil {
			panic(fmt.Errorf("launch playwrite: %w", err))
		}

		browser, err := pw.Chromium.Launch()
		if err != nil {
			panic(fmt.Errorf("launch Chromium: %w", err))
		}

		pwContext, err = browser.NewContext()
		if err != nil {
			panic(fmt.Errorf("playwrite browser context: %w", err))
		}
	})

	page, err := pwContext.NewPage()
	if err != nil {
		panic(fmt.Errorf("playwrite context new page: %w", err))
	}

	return page
}
