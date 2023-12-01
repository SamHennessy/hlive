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
			file         = l.NewLockBox(l.File{})
			fileName     = l.Box("")
			fileType     = l.Box("")
			fileSize     = l.Box(0)
			tableDisplay = l.NewLockBox("none")
			fileDisplay  = l.NewLockBox("none")
		)

		page := l.NewPage()
		page.DOM().Title().Add("File Upload Example")
		page.DOM().Head().Add(l.T("link",
			l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		iframe := l.T("iframe", l.Style{"width": "100%", "height": "80vh"})

		fileInput := l.C("input", l.Attrs{"type": "file"})
		fileInput.Add(
			l.On("upload", func(ctx context.Context, e l.Event) {
				file.Lock(func(v l.File) l.File {
					// This is a bad idea, using as an easy demo.
					// The file is kept in server-side browser DOM using memory.
					src := fmt.Sprintf("data:%s;base64,%s", e.File.Type,
						base64.StdEncoding.EncodeToString(e.File.Data))
					iframe.Add(l.Attrs{"src": src})
					fileDisplay.Set("box")
					fileInput.RemoveAttributes(l.AttrUpload)

					return *e.File
				})
			}),
			l.On("change", func(ctx context.Context, e l.Event) {
				if e.File != nil {
					file.Lock(func(v l.File) l.File {
						v = *e.File
						fileName.Set(v.Name)
						fileType.Set(v.Type)
						fileSize.Set(v.Size)

						tableDisplay.Set("box")

						return v
					})
				}
			}),
		)

		uploadBtn := l.C("button", "Upload",
			l.On("click", func(ctx context.Context, e l.Event) {
				fileInput.Add(l.Attrs{l.AttrUpload: ""})
			}),
		)

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "Upload"),
				l.T("p", "Example of using the file upload features."),
			),
			l.T("main",
				l.T("p",
					fileInput,
					uploadBtn,
				),
				l.T("div", l.StyleLockBox{"display": tableDisplay},
					l.T("table",
						l.T("tbody",
							l.T("tr",
								l.T("td", "Name"), l.T("td", fileName),
							),
							l.T("tr",
								l.T("td", "Type"), l.T("td", fileType),
							),
							l.T("tr",
								l.T("td", "Size"), l.T("td", l.T("i", fileSize), " bytes"),
							),
						),
					),
				),
				l.T("div", l.StyleLockBox{"display": fileDisplay},
					l.T("h3", "Uploaded File"),
					l.T("hr"),
					iframe,
				),
			),
		)

		return page
	}

	return l.NewPageServer(f)
}
