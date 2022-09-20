package main

import (
	"context"
	"log"
	"net/http"
	"time"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("Clock Example")
		page.DOM().Head().Add(
			l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "Clock"),
				l.T("p", "The time updates are being push from the server every 10ms"),
			),
			l.T("main",
				l.T("pre", newClock()),
			),
		)

		return page
	}

	ps := l.NewPageServer(f)
	// Kill the page session 1 second after the tab is closed
	ps.Sessions.DisconnectTimeout = time.Second

	return ps
}

func newClock() *clock {
	t := l.NewLockBox("")

	return &clock{
		Component: l.C("code", "Server: ", t),
		timeStr:   t,
	}
}

type clock struct {
	*l.Component

	timeStr *l.LockBox[string]
	tick    *time.Ticker
}

func (c *clock) Mount(ctx context.Context) {
	log.Println("DEBU: start tick")

	c.tick = time.NewTicker(10 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("DEBU: tick loop stop: ctx")

				return
			case t := <-c.tick.C:
				c.timeStr.Set(t.String())

				l.RenderComponent(ctx, c)
			}
		}
	}()
}

// Unmount
// Will be called after the page session is deleted
func (c *clock) Unmount(_ context.Context) {
	log.Println("DEBU: stop tick")

	if c.tick != nil {
		c.tick.Stop()
	}
}
