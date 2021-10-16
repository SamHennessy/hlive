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
		page.Title.Add("Clock Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(

			l.T("h1", "Clock"),
			l.T("blockquote", "The time updates are being push from the server every 100ms"),
			l.T("pre", newClock()),
		)

		return page
	}

	ps := l.NewPageServer(f)
	// Still kill the page session 1 second after the tab is closed
	ps.Sessions.DisconnectTimeout = time.Second

	return ps
}

func newClock() *clock {
	return &clock{
		Component: l.C("code"),
		t:         time.Now(),
	}
}

type clock struct {
	*l.Component

	t    time.Time
	tick *time.Ticker
	done chan bool
}

func (c *clock) GetNodes() *l.NodeGroup {
	return l.G("Server Time: " + c.t.String())
}

func (c *clock) Mount(ctx context.Context) {
	log.Println("DEBU: start tick")

	c.tick = time.NewTicker(100 * time.Millisecond)
	c.done = make(chan bool)

	go func() {
		for {
			select {
			case <-c.done:
				log.Println("DEBU: tick loop stop")

				return
			case t := <-c.tick.C:
				c.t = t

				l.RenderComponent(ctx, c)
			}
		}
	}()
}

// Unmount
// Will be called after the page session is deleted
func (c *clock) Unmount(_ context.Context) {
	log.Println("DEBU: stop tick")
	c.done <- true

	if c.tick != nil {
		c.tick.Stop()
	}
}
