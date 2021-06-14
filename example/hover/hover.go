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
		hoverState := ""

		hover := l.C("span",
			l.OnMouseEnter(func(ctx context.Context, e l.Event) {
				hoverState = "Mouse enter"
			}),
			l.OnMouseLeave(func(ctx context.Context, e l.Event) {
				hoverState = "Mouse leave"
			}),
			"Hover over me",
		)

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Hover Example")
		page.Body.Add(
			l.T("div", hover),
			l.T("div", &hoverState),
		)

		return page
	}

	return l.NewPageServer(f)
}
