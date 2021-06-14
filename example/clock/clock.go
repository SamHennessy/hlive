package main

import (
	"context"
	"net/http"
	"os"
	"time"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

// TODO: allow user to adjust tick duration, and allow pausing of the clock

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	http.Handle("/", Home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func Home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Clock Example")
		page.Body.Add(Clock(logger))

		return page
	}

	ps := l.NewPageServer(f)
	ps.Sessions.DisconnectTimeout = time.Second

	return ps
}

func Clock(logger zerolog.Logger) *clock {
	return &clock{
		Component: l.C("span"),
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

func (c *clock) GetNodes() []interface{} {
	return l.Tree(c.t.String())
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

func (c *clock) Unmount(ctx context.Context) {
	c.logger.Info().Msg("stop tick")
	c.done <- true
	if c.tick != nil {
		c.tick.Stop()
	}
}
