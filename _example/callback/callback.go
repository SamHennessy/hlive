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

func callback(container *l.Component) {
	container.Add(l.On(l.DiffApply, func(ctx context.Context, e l.Event) {
		container.Add(l.T("p", "Diff Apply"))
		container.RemoveEventBinding(e.Binding.ID)
	}))
}

func home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		container := l.C("div")

		btn := l.C("button", "Trigger Click",
			l.On("click", func(ctx context.Context, e l.Event) {
				container.Add(l.T("p", "Click"))
				callback(container)
			}),
		)

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Callback Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(btn, l.T("h1", "Events"), container)

		return page
	}

	return l.NewPageServer(f)
}
