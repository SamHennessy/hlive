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

	http.Handle("/", home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		itemList := List()
		in := newTextInput()
		in.On(l.OnKeyUp(func(ctx context.Context, e l.Event) {
			// We need to allow 1 render so we can clear later
			doRender := in.Value == ""
			newVal := e.Value

			if e.Key == "Enter" && in.Value != "" {
				itemList.items = append(itemList.items, newTextItem(itemList, in.Value))
				newVal = ""
				doRender = true
			}

			in.Value = newVal

			if doRender {
				l.RenderWS(ctx)
			}
		}))

		btn := newButton("Add")
		btn.On(l.OnClick(func(ctx context.Context, e l.Event) {
			if in.Value != "" {
				itemList.items = append(itemList.items, newTextItem(itemList, in.Value))
			}

			in.Value = ""
		}))

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Add Remove Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))
		page.Body.Add(
			in,
			btn,
			itemList,
		)
		//
		return page
	}

	return l.NewPageServer(f)
}

func newTextInput() *textInput {
	c := &textInput{
		Component: l.NewComponent("input", l.Attrs{"type": "text", "placeholder": "Label"}),
	}

	c.AutoRender = false

	c.On(l.OnFocus(func(ctx context.Context, e l.Event) {
		c.Focus = false
	}))

	return c
}

type textInput struct {
	*l.Component

	Value string
	Focus bool
}

func (c *textInput) GetNodes() []interface{} {
	c.SetAttributes(l.Attrs{"value": c.Value})
	if c.Focus {
		c.SetAttributes(l.Attrs{l.AttrFocus: nil})
	} else {
		c.RemoveAttributes(l.AttrFocus)
	}

	return nil
}

func newButton(label string) *button {
	c := &button{
		Component: l.C("button"),
		Label:     label,
	}

	return c
}

type button struct {
	*l.Component

	Label string
}

func (c *button) GetNodes() []interface{} {
	c.SetAttributes(l.Attrs{"type": "button"})

	return l.Tree(c.Label)
}

func List() *itemList {
	return &itemList{
		Component: l.C("div"),
	}
}

type itemList struct {
	*l.Component

	items []l.Componenter
}

func (c *itemList) GetNodes() []interface{} {
	var list []interface{}
	for i := 0; i < len(c.items); i++ {
		list = append(list, c.items[i])
	}

	return list
}

func (c *itemList) DeleteItem(item l.Componenter) {
	var newItems []l.Componenter

	for i := 0; i < len(c.items); i++ {
		if c.items[i] == item {
			continue
		}

		newItems = append(newItems, c.items[i])
	}

	c.items = newItems
}

func newTextItem(list *itemList, label string) *textItem {
	item := &textItem{
		Component: l.C("div"),
		Label:     label,
	}

	in := newTextInput()
	in.On(l.OnKeyUp(func(ctx context.Context, e l.Event) {
		doRender := in.Value == ""
		newVal := e.Value

		if e.Key == "Enter" {
			item.Label = in.Value
			item.EditMode = false
			doRender = true
		}

		in.Value = newVal

		if doRender {
			l.RenderWS(ctx)
		}
	}))

	item.Input = in

	edit := newClickLink("Edit")
	edit.On(l.OnClick(func(ctx context.Context, e l.Event) {
		in.Value = item.Label
		in.Focus = true
		item.EditMode = true
	}))

	item.EditLink = edit

	can := newClickLink("Cancel")
	can.On(l.OnClick(func(ctx context.Context, e l.Event) {
		item.EditMode = false
	}))

	item.CancelLink = can

	del := newClickLink("Delete")
	del.On(l.OnClick(func(ctx context.Context, e l.Event) {
		list.DeleteItem(item)
	}))

	item.DeleteLink = del

	btn := newButton("Update")
	btn.On(l.OnClick(func(ctx context.Context, e l.Event) {
		item.Label = in.Value
		in.Value = ""
		item.EditMode = false
	}))

	item.UpdateBtn = btn

	return item
}

type textItem struct {
	*l.Component

	Label      string
	EditMode   bool
	EditLink   l.Componenter
	DeleteLink l.Componenter
	CancelLink l.Componenter
	UpdateBtn  l.Componenter
	Input      l.Componenter
}

func (c *textItem) GetNodes() []interface{} {
	var kids []interface{}

	if c.EditMode {
		kids = []interface{}{
			c.Input, " ",
			c.UpdateBtn, " ",
			c.CancelLink,
		}
	} else {
		kids = []interface{}{
			c.DeleteLink,
			" ", c.EditLink,
			" " + c.Label,
		}
	}

	return kids
}

func newClickLink(children ...interface{}) *clickLink {
	// A child could overwrite these attributes
	children = append([]interface{}{l.Attrs{"href": "#"}}, children...)
	c := &clickLink{
		Component: l.C("a", children...),
	}

	return c
}

type clickLink struct {
	*l.Component
}
