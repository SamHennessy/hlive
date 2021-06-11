package main

import (
	"context"
	"net/http"
	"os"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

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
		count := 0

		page := l.NewPage()
		page.Logger = logger
		page.Title.Add("Local Render Example")
		page.Body.Add(
			l.T("p", "Global Render"),
			CountBtn(&count),
			l.Tree("The count is: ", l.T("em", &count), " clicks"),
			l.T("hr"),
			l.T("p", "Local Render"),
			CountBtnLocal(&count),
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

func (c *countBtn) Render() []interface{} {
	return l.Tree(c.Count)
}

func CountBtn(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}

	c.On(l.OnClick(func(ctx context.Context, e l.Event) {
		*c.Count++
	}))

	return c
}

func CountBtnLocal(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}
	c.AutoRender = false

	c.On(l.OnClick(func(ctx context.Context, e l.Event) {
		*c.Count++

		l.RenderComponentWS(ctx, c)
	}))

	return c
}
