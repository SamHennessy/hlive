package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/",
		urlParamsMiddleware(
			l.NewPageServer(home(logger)).ServeHTTP,
		),
	)

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("URL Params Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(l.T("h1", "URL Get Parameter Read Example"),
			l.T("p", "This example reads the parameters from the URL and prints them in a table. "+
				"You will see the extra 'hlive' parameter that HLive adds on when establishing a WebSocket connection."),
			l.T("p", "Add your own query parameters to the url and load the page again."),
		)

		cl := l.List("tbody")

		cm := l.CM("table",
			l.T("thead",
				l.T("tr",
					l.T("th", "Key"),
					l.T("th", "Value"),
				),
			),
			cl,
		)

		cm.MountFunc = func(ctx context.Context) {
			for key, value := range urlParamsFromCtx(ctx) {
				cl.AddItem(l.CM("tr",
					l.T("td", key),
					l.T("td", value),
				))
			}
		}

		page.Body.Add(cm)

		return page
	}
}

type ctxKey string

const ctxURLParams ctxKey = "url"

func urlParamsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only on WebSocket requests
		if r.URL.Query().Get("hlive") != "" {
			params := map[string]string{}
			for key := range r.URL.Query() {
				params[key] = r.URL.Query().Get(key)
			}

			r = r.WithContext(context.WithValue(r.Context(), ctxURLParams, params))
		}

		h(w, r)
	}
}

func urlParamsFromCtx(ctx context.Context) map[string]string {
	params, ok := ctx.Value(ctxURLParams).(map[string]string)
	if ok && params != nil {
		return params
	}

	return map[string]string{}
}
