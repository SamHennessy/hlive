package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	http.Handle("/", l.NewPageServer(home(logger)))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) func() *l.Page {
	return func() *l.Page {
		var (
			file         l.File
			tableDisplay = "none"
			fileDisplay  = "none"
		)

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("File Upload Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		iframe := l.T("iframe", l.Style{"width": "100%", "height": "80vh"})

		fileInput := l.C("input", l.Attrs{"type": "file"})
		fileInput.Add(
			l.OnUpload(func(ctx context.Context, e l.Event) {
				fileDisplay = "box"
				fileInput.RemoveAttributes(l.AttrUpload)
				// This is a bad idea, only using as an easy demo.
				// The reason is it will be in the page tree using a lot of memory
				src := fmt.Sprintf("data:%s;base64,%s", e.File.Type, base64.StdEncoding.EncodeToString(e.File.Data))

				iframe.Add(l.Attrs{"src": src})
			}),
			l.OnChange(func(ctx context.Context, e l.Event) {
				if e.File != nil {
					tableDisplay = "box"
					file = *e.File
				}
			}),
		)

		uploadBtn := l.C("button", "Upload",
			l.OnClick(func(ctx context.Context, e l.Event) {
				fileInput.Add(l.Attrs{l.AttrUpload: ""})
			}),
		)

		page.Body.Add(
			l.T("h1", "Upload"),
			l.T("hr"),
			fileInput,
			uploadBtn,
			l.T("div", l.Style{"display": &tableDisplay},
				l.T("table",
					l.T("tbody",
						l.T("tr",
							l.T("td", "Name"), l.T("td", &file.Name),
						),
						l.T("tr",
							l.T("td", "Type"), l.T("td", &file.Type),
						),
						l.T("tr",
							l.T("td", "Size"), l.T("td", l.T("i", &file.Size), " bytes"),
						),
					),
				),
			),
			l.T("div", l.Style{"display": &fileDisplay},
				l.T("h3", "Uploaded File"),
				l.T("hr"),
				iframe,
			),
		)

		return page
	}
}
