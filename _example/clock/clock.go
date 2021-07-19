package main

import (
	"context"
	"net/http"
	"time"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

// Feature Ideas: allow user to adjust tick duration, and allow pausing of the clock

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/", home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Clock Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(l.T("pre", newClock(logger)))

		return page
	}

	ps := l.NewPageServer(f)
	// Still kill the page session 1 second after the tab is closed
	ps.Sessions.DisconnectTimeout = time.Second

	return ps
}

func newClock(logger zerolog.Logger) *clock {
	return &clock{
		Component: l.C("code"),
		logger:    logger,
		t:         time.Now(),
	}
}

type clock struct {
	*l.Component

	logger zerolog.Logger
	t      time.Time
	tick   *time.Ticker
	done   chan bool
}

func (c *clock) GetNodes() interface{} {
	return "Server Time: " + c.t.String()
}

func (c *clock) Mount(ctx context.Context) {
	c.logger.Info().Msg("start tick")
	c.tick = time.NewTicker(time.Second / 10)

	c.done = make(chan bool)

	go func() {
		for {
			select {
			case <-c.done:
				c.logger.Info().Msg("tick loop stop")

				return
			case t := <-c.tick.C:
				c.t = t

				l.RenderWS(ctx)
			}
		}
	}()
}

// Unmount
// Will be called after the page session is deleted
func (c *clock) Unmount(_ context.Context) {
	c.logger.Info().Msg("stop tick")
	c.done <- true

	if c.tick != nil {
		c.tick.Stop()
	}
}
