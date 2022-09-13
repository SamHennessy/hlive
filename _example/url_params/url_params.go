package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivekit"
)

func main() {
	http.Handle("/",
		urlParamsMiddleware(
			home().ServeHTTP,
		),
	)

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("URL Params Example")
		page.DOM().Head().Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "URL Get Parameter Read Example"),
				l.T("p", "This example reads the parameters from the URL and prints them in a table."),
			),
			l.T("main",
				l.T("p", "Add your own query parameters to the url and load the page again."),
				l.T("h2", "Values"),
			),
		)

		cl := hlivekit.List("tbody")

		cm := l.CM("table",
			l.T("thead",
				l.T("tr",
					l.T("th", "Key"),
					l.T("th", "Value"),
				),
			),
			cl,
		)

		cm.mountFunc = func(ctx context.Context) {
			for key, value := range urlParamsFromCtx(ctx) {
				cl.AddItem(l.CM("tr",
					l.T("td", key),
					l.T("td", value),
				))
			}
		}

		page.DOM().Body().Add(
			cm,
			l.T("p", "You will see the extra 'hlive' parameter that HLive adds on when establishing a WebSocket connection."),
		)

		return page
	}

	return l.NewPageServer(f)
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
