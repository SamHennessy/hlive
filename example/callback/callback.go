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

func callback(container *l.Component) {
	container.On(l.OnDiffApply(func(ctx context.Context, e l.Event) {
		container.Add(l.T("p", "OnDiffApply"))
		container.RemoveEventBinding(e.Binding.ID)
	}))
}

func Home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		container := l.C("div")

		btn := l.C("button", "Trigger OnClick",
			l.OnClick(func(ctx context.Context, e l.Event) {
				container.Add(l.T("p", "OnClick"))
				callback(container)
			}),
		)

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Callback Example")
		page.Body.Add(btn, l.T("h1", "Events"), container)

		return page
	}

	return l.NewPageServer(f)
}
