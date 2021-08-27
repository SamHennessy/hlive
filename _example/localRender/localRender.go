package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

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
		count := 0

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Local GetNodes Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("h2", "Global Render"),
			newCountBtn(&count),
			l.Tree("The count is: ", l.T("em", &count), " clicks"),
			l.T("h2", "Local Render"),
			newCountBtnLocal(&count),
			l.Tree("The count is: ", l.T("em", &count), " clicks"),
		)

		return page
	}

	return l.NewPageServer(f)
}

type countBtn struct {
	*l.Component

	Count *int
}

func (c *countBtn) GetNodes() interface{} {
	return l.Tree(c.Count)
}

func newCountBtn(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		*c.Count++
	}))

	return c
}

func newCountBtnLocal(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}

	// Don't render this component when an event binding is triggered
	c.AutoRender = false

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		*c.Count++

		// Will render the passed component and it's subtree
		l.RenderComponentWS(ctx, c)
	}))

	return c
}
