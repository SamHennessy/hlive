package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

const pageStyle l.HTML = `
.box {
	display: grid;
	grid-template-columns: 1fr 1fr 1fr;
	gap: 0.5em;
}

input, select {
	width: 100%;
}
`

func home() *l.Page {
	page := l.NewPage()
	page.DOM().Title().Add("Form Data Initial Sync Example")
	page.DOM().Head().Add(l.T("link",
		l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))
	page.DOM().Head().Add(l.T("style", pageStyle))

	var (
		formValsSync   [9]*l.NodeBox[string]
		formValsNoSync [len(formValsSync)]*l.NodeBox[string]
	)

	for i := 0; i < len(formValsSync); i++ {
		formValsSync[i] = l.Box("")
		formValsNoSync[i] = l.Box("")
	}

	page.DOM().Body().Add(
		l.T("header",
			l.T("h1", "Form Data Initial Sync"),
			l.T("p", "Some browsers, life FireFox, don't clear form field data after a page reload. "+
				"HLive will send this data to relevant inputs when this happens."),
		),
		l.T("main",
			l.T("p", "To test, using FireFox, change the fields below then reload."),
			l.T("h2", "Form"),
			l.T("form", l.Class("box"),
				//
				// Text
				//
				l.T("div",
					l.T("label", "Text"),
					l.T("br"),
					l.C("input", l.Attrs{"type": "text"},
						l.On("input", func(_ context.Context, e l.Event) {
							if !e.IsInitial {
								formValsNoSync[0].Set(e.Value)
							}

							formValsSync[0].Set(e.Value)
						}),
					),
				),
				//
				// Password
				//
				l.T("div",
					l.T("label", "Password"),
					l.T("br"),
					l.C("input", l.Attrs{"type": "password"},
						l.On("input", func(_ context.Context, e l.Event) {
							if !e.IsInitial {
								formValsNoSync[1].Set(e.Value)
							}

							formValsSync[1].Set(e.Value)
						})),
				),
				//
				// Range
				//
				l.T("div",
					l.T("label", "Range"),
					l.T("br"),
					l.C("input", l.Attrs{"type": "range", "min": "0", "max": "1000"},
						l.On("input", func(_ context.Context, e l.Event) {
							if !e.IsInitial {
								formValsNoSync[2].Set(e.Value)
							}

							formValsSync[2].Set(e.Value)
						}),
					),
				),
				//
				// Multi select
				//
				l.T("div",
					l.T("label", "Multi Select"),
					l.T("br"),
					l.C("select", l.Attrs{"multiple": ""},
						l.On("input", func(_ context.Context, e l.Event) {
							if !e.IsInitial {
								formValsNoSync[3].Set(strings.Join(e.Values, ", "))
							}

							formValsSync[3].Set(strings.Join(e.Values, ", "))
						}),
						l.T("option", l.Attrs{"value": "dog"}, "Dog"),
						l.T("option", l.Attrs{"value": "cat"}, "Cat"),
						l.T("option", l.Attrs{"value": "bird"}, "Bird"),
						l.T("option", "Fox"),
					),
					l.T("br"),
					l.T("small", "Click + Ctl/Cmd to multi select"),
				),
				//
				// Radio
				//
				l.T("div",
					l.T("label", "Radio"),
					l.T("br"),
					l.T("label", l.Attrs{"for": "radio_1"},
						l.C("input",
							l.Attrs{"type": "radio", "name": "radio", "value": "orange", "id": "radio_1"},
							l.On("input", func(_ context.Context, e l.Event) {
								if !e.IsInitial {
									formValsNoSync[4].Set(e.Value)
								}

								formValsSync[4].Set(e.Value)
							}),
						),
						" Orange"),
					l.T("label", l.Attrs{"for": "radio_2"},
						l.C("input",
							l.Attrs{"type": "radio", "name": "radio", "value": "grape", "id": "radio_2"},
							l.On("input", func(_ context.Context, e l.Event) {
								if !e.IsInitial {
									formValsNoSync[4].Set(e.Value)
								}

								formValsSync[4].Set(e.Value)
							}),
						),
						" Grape"),
					l.T("label", l.Attrs{"for": "radio_3"},
						l.C("input",
							l.Attrs{"type": "radio", "name": "radio", "value": "lemon", "id": "radio_3"},
							l.On("input", func(_ context.Context, e l.Event) {
								if !e.IsInitial {
									formValsNoSync[4].Set(e.Value)
								}

								formValsSync[4].Set(e.Value)
							}),
						),
						" Lemon"),
					l.T("label", l.Attrs{"for": "radio_4"},
						l.C("input",
							l.Attrs{"type": "radio", "name": "radio", "value": "apple", "id": "radio_4"},
							l.On("input", func(_ context.Context, e l.Event) {
								if !e.IsInitial {
									formValsNoSync[4].Set(e.Value)
								}

								formValsSync[4].Set(e.Value)
							}),
						),
						" Apple"),
				),
				//
				// Checkbox
				//
				l.T("div",
					l.T("label", "Checkbox"),
					l.T("br"),

					l.T("label", l.Attrs{"for": "checkbox_1"},
						l.C("input", l.Attrs{"type": "checkbox", "value": "north", "id": "checkbox_1"},
							l.On("input", func(_ context.Context, e l.Event) {
								formValsNoSync[5].Set("")
								formValsSync[5].Set("")

								if !e.IsInitial && e.Selected {
									formValsNoSync[5].Set(e.Value)
								}

								if e.Selected {
									formValsSync[5].Set(e.Value)
								}
							}),
						),
						"North"),
					l.T("label", l.Attrs{"for": "checkbox_2"},
						l.C("input", l.Attrs{"type": "checkbox", "value": "east", "id": "checkbox_2"},
							l.On("input", func(_ context.Context, e l.Event) {
								formValsNoSync[6].Set("")
								formValsSync[6].Set("")

								if !e.IsInitial && e.Selected {
									formValsNoSync[6].Set(e.Value)
								}

								if e.Selected {
									formValsSync[6].Set(e.Value)
								}
							}),
						),
						"East"),
					l.T("label", l.Attrs{"for": "checkbox_3"},
						l.C("input", l.Attrs{"type": "checkbox", "value": "south", "id": "checkbox_3"},
							l.On("input", func(_ context.Context, e l.Event) {
								formValsNoSync[7].Set("")
								formValsSync[7].Set("")

								if !e.IsInitial && e.Selected {
									formValsNoSync[7].Set(e.Value)
								}

								if e.Selected {
									formValsSync[7].Set(e.Value)
								}
							}),
						),
						"South"),
					l.T("label", l.Attrs{"for": "checkbox_4"},
						l.C("input", l.Attrs{"type": "checkbox", "value": "west", "id": "checkbox_4"},
							l.On("input", func(_ context.Context, e l.Event) {
								formValsNoSync[8].Set("")
								formValsSync[8].Set("")

								if !e.IsInitial && e.Selected {
									formValsNoSync[8].Set(e.Value)
								}

								if e.Selected {
									formValsSync[8].Set(e.Value)
								}
							}),
						),
						"West"),
				),
			),
			//
			// Output
			//
			l.T("h2", "Server Side Data"),
			l.T("table",
				l.T("thead",
					l.T("tr",
						l.T("th", ""),
						l.T("th", "Sync"),
						l.T("th", "No Sync"),
					),
				),
				l.T("tbody",
					l.T("tr",
						l.T("td", "Text"),
						l.T("td", formValsSync[0]),
						l.T("td", formValsNoSync[0]),
					),
					l.T("tr",
						l.T("td", "Password", l.T("br"),
							l.T("small", "Browsers won't keep this on refresh")),
						l.T("td", formValsSync[1]),
						l.T("td", formValsNoSync[1]),
					),
					l.T("tr",
						l.T("td", "Range", l.T("br"),
							l.T("small", "Browsers set this to the mid point by default")),
						l.T("td", formValsSync[2]),
						l.T("td", formValsNoSync[2]),
					),
					l.T("tr",
						l.T("td", "Multi Select"),
						l.T("td", formValsSync[3]),
						l.T("td", formValsNoSync[3]),
					),
					l.T("tr",
						l.T("td", "Radio"),
						l.T("td", formValsSync[4]),
						l.T("td", formValsNoSync[4]),
					),
					l.T("tr",
						l.T("td", "Checkbox"),
						l.T("td",
							formValsSync[5], " ", formValsSync[6], " ", formValsSync[7], " ", formValsSync[8]),
						l.T("td",
							formValsNoSync[5], " ", formValsNoSync[6], " ", formValsNoSync[7], " ", formValsNoSync[8]),
					),
				),
			),
		),
	)

	return page
}
