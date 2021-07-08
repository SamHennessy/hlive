package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/", Home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func Home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		hoverState := " "

		hover := l.C("h1",
			l.On("mouseEnter", func(ctx context.Context, e l.Event) {
				hoverState = "Mouse enter"
			}),
			l.On("mouseLeave", func(ctx context.Context, e l.Event) {
				hoverState = "Mouse leave"
			}),
			"Hover over me",
		)

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Hover Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("div", hover),
			l.T("pre", l.T("code", &hoverState)),
		)

		return page
	}

	return l.NewPageServer(f)
}
