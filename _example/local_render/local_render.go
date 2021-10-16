package main

import (
	"context"
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
		count := 0

		page := l.NewPage()
		page.Title.Add("Local GetNodes Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("h1", "Local Render"),
			l.T("blockquote", "By default, the whole page if checked for differance after an event. "+
				"You can override that behaviour and chose to only render a component and it's children."),
			l.T("h2", "Global Render"),
			l.T("h4", "Everything will update"),
			newCountBtn(&count),
			l.Group("The count is: ", l.T("em", &count), " clicks"),
			l.T("h2", "Local Render"),
			l.T("h4", "Only the button will update"),
			newCountBtnLocal(&count),
			l.Group("The count is: ", l.T("em", &count), " clicks"),
		)

		return page
	}

	return l.NewPageServer(f)
}

type countBtn struct {
	*l.Component

	Count *int
}

func (c *countBtn) GetNodes() *l.NodeGroup {
	return l.G(c.Count)
}

func newCountBtn(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		*c.Count++
	}))

	return c
}

func newCountBtnLocal(count *int) *countBtn {
	c := &countBtn{
		Component: l.C("button"),
		Count:     count,
	}

	// Don't render this component when an event binding is triggered
	c.AutoRender = false

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		*c.Count++

		// Will render the passed component and it's subtree
		l.RenderComponent(ctx, c)
	}))

	return c
}
