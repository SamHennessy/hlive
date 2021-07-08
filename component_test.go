package hlive_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

type countBtn struct {
	*hlive.Component

	Count int
}

func CountBtn() *countBtn {
	return &countBtn{
		Component: hlive.C("span"),
	}
}

func (c *countBtn) GetNodes() interface{} {
	return hlive.NewTag("button", c.Count)
}

func TestPage_RenderElementComponent(t *testing.T) {
	t.Parallel()

	el := hlive.C("span")
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	expect := fmt.Sprintf(`<span data-hlive-id="%s"></span>`, el.GetID())

	if diff := deep.Equal(expect, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func pageToDoc(t *testing.T, ctx context.Context, page *hlive.Page) *goquery.Document {
	t.Helper()

	buff := bytes.NewBuffer(nil)

	if err := page.RenderHTML(ctx, buff); err != nil {
		t.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(buff)
	if err != nil {
		t.Fatal("goquery:", err)
	}

	return doc
}

func TestPage_ComponentEventOnClickRender(t *testing.T) {
	t.Parallel()

	page := hlive.NewPage()
	comp := hlive.C("span")
	handler := func(ctx context.Context, e hlive.Event) {}

	comp.Add(hlive.Attrs{"id": "content"}, hlive.On("click", handler))

	page.Body.Add(comp)

	ctx := context.Background()
	ctx = hlive.SetIsWebSocket(ctx)

	doc := pageToDoc(t, ctx, page)

	val, _ := doc.Find("#content").Attr(hlive.AttrID)
	if diff := deep.Equal(comp.GetID(), val); diff != nil {
		t.Error(diff)
	}

	val, exists := doc.Find("#content").Attr(hlive.AttrOn)
	if !exists {
		t.Error(hlive.AttrOn + " attribute not found ")
	} else {
		// TODO: fix
		// parts := strings.Split(val, "|")
		// a := comp.GetEventBinding(parts[1])
		//
		// if diff := deep.Equal(hlive.EventHandler(handler), a.Handler); diff != nil {
		// 	t.Error(diff)
		// }
	}
}

// TODO: make work
// func TestPage_ComponentEventOnClick(t *testing.T) {
// 	t.Parallel()
//
// 	// Create counting button component
// 	el := CountBtn()
// 	el.Add(hlive.Attrs{"id": "content"})
// 	// Assign a handler that will increment it's counter
// 	el.Add(hlive.On("click", func(ctx context.Context, e hlive.Event) {
// 		el.Count++
// 	}))
// 	// Add to a page
// 	page := hlive.NewPage()
// 	page.Body.Add(el)
// 	// HTML page
// 	ctx := hlive.SetIsWebSocket(context.Background())
// 	doc := pageToDoc(t, ctx, page)
// 	// TODO: I don't need to get this from the page any more but maybe I should?
// 	// Get the binding id
// 	val, exists := doc.Find("#content").Attr(hlive.AttrOn)
// 	if !exists {
// 		t.Fatal(hlive.AttrOn + " attribute not found ")
// 	}
// 	// Get the binding
// 	binding := el.GetEventBinding(val)
// 	// Test data
// 	if diff := deep.Equal("0", doc.Find("#content").First().Text()); diff != nil {
// 		t.Error(diff)
// 	}
// 	// Call handler, increment counter
// 	binding.Handler(context.Background(), hlive.Event{})
// 	// Rerender
// 	// HTML page
// 	doc = pageToDoc(t, context.Background(), page)
// 	// Test data again
// 	if diff := deep.Equal("1", doc.Find("#content").First().Text()); diff != nil {
// 		t.Error(diff)
// 	}
// }
