package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		var (
			file         l.File
			tableDisplay = "none"
			fileDisplay  = "none"
		)

		page := l.NewPage()
		page.Title.Add("File Upload Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		iframe := l.T("iframe", l.Style{"width": "100%", "height": "80vh"})

		fileInput := l.C("input", l.Attrs{"type": "file"})
		fileInput.Add(
			l.On("upload", func(ctx context.Context, e l.Event) {
				fileDisplay = "box"
				fileInput.RemoveAttributes(l.AttrUpload)
				// This is a bad idea, only using as an easy demo.
				// The reason it's a bad idea is it will be in the server page tree using a lot of memory
				src := fmt.Sprintf("data:%s;base64,%s", e.File.Type, base64.StdEncoding.EncodeToString(e.File.Data))

				iframe.Add(l.Attrs{"src": src})
			}),
			l.On("change", func(ctx context.Context, e l.Event) {
				if e.File != nil {
					tableDisplay = "box"
					file = *e.File
				}
			}),
		)

		uploadBtn := l.C("button", "Upload",
			l.On("click", func(ctx context.Context, e l.Event) {
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

	return l.NewPageServer(f)
}
