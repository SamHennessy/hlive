package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/", l.NewPageServer(home(logger)))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Click Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		var count int

		page.Body.Add(
			"Clicks: ",
			l.C("button",
				l.On("click", func(_ context.Context, _ l.Event) {
					count++
				}),
				// Passing by reference
				&count,
			),
		)

		return page
	}
}
