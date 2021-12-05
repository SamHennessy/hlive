package systemtests_test

import (
	"context"
	"testing"
	"time"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

func TestUnmount_CloseTab(t *testing.T) {
	t.Parallel()

	done := make(chan bool)

	c := l.CM("div")
	c.UnmountFunc = func(ctx context.Context) {
		done <- true
	}

	page := l.NewPage()

	page.DOM.Body.Add(c)

	pageFn := func() *l.Page {
		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	_, err := h.pwpage.WaitForFunction("hlive.sessID != 1", nil)

	hlivetest.FatalOnErr(t, err)

	// No wait after disconnect
	h.server.PageSessionStore.DisconnectTimeout = 0

	hlivetest.FatalOnErr(t, h.pwpage.Close())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-done:
		return
	case <-ctx.Done():
		t.Error("timed out waiting for unmount")
	}
}
