package pages_test

import (
	"testing"

	"github.com/SamHennessy/hlive/hlivetest"
	"github.com/SamHennessy/hlive/hlivetest/pages"
)

func TestClick_OneClick(t *testing.T) {
	t.Parallel()

	h := setup(t, pages.Click())
	defer h.teardown()

	hlivetest.Diff(t, "0", hlivetest.TextContent(t, h.pwpage, "#count"))

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "1", hlivetest.TextContent(t, h.pwpage, "#count"))
}

func TestClick_TenClick(t *testing.T) {
	t.Parallel()

	h := setup(t, pages.Click())
	defer h.teardown()

	hlivetest.Diff(t, "0", hlivetest.TextContent(t, h.pwpage, "#count"))

	for i := 0; i < 9; i++ {
		hlivetest.Click(t, h.pwpage, "#btn")
	}

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "10", hlivetest.TextContent(t, h.pwpage, "#count"))
}
