package main

import (
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

// Step 1
func home() *l.Page {
	page := l.NewPage()
	page.DOM().Body().Add("Hello, world.")

	return page
}

// Step 2
// func home() *l.Page {
// 	var message string
//
// 	input := l.NewComponent("input")
// 	input.Add(l.Attrs{"type": "text"})
// 	input.Add(l.On("keyup", func(ctx context.Context, e l.Event) {
// 		message = e.Value
// 	}))
//
// 	page := l.NewPage()
// 	page.DOM().Body().Add(l.NewTag("div", input))
// 	page.DOM().Body().Add(l.T("hr"))
// 	page.DOM().Body().Add("Hello, ", &message)
//
// 	return page
// }

// Step 2.1
// func home() *l.Page {
// 	var message string
//
// 	input := l.C("input",
// 		l.Attrs{"type": "text"},
// 		l.On("keyup", func(_ context.Context, e l.Event) {
// 			message = e.Value
// 		}),
// 	)
//
// 	page := l.NewPage()
// 	page.DOM().Body().Add(
// 		l.T("div", input),
// 		l.T("hr"),
// 		"Hello, ", &message,
// 	)
//
// 	return page
// }

func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("Listing on :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("Error: http listen and serve:", err)
	}
}
