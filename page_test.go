package hlive_test

import (
	"bytes"
	"testing"

	"github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestPage_RenderElementText(t *testing.T) {
	t.Parallel()

	buff := bytes.NewBuffer(nil)
	if err := hlive.NewRender().HTML(buff, "<h1>text_test</h1>"); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("&lt;h1&gt;text_test&lt;/h1&gt;", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementRawHTML(t *testing.T) {
	t.Parallel()

	el := hlive.HTML("<h1>html_test</h1>")
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("<h1>html_test</h1>", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTag(t *testing.T) {
	t.Parallel()

	el := hlive.NewTag("a")
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("<a></a>", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagAttr(t *testing.T) {
	t.Parallel()

	el := hlive.NewTag("a", hlive.NewAttribute("href", "https://example.com"))
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a href="https://example.com"></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagAttrs(t *testing.T) {
	t.Parallel()

	el := hlive.NewTag("a", hlive.Attrs{"href": "https://example.com"})
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a href="https://example.com"></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagChildText(t *testing.T) {
	t.Parallel()

	el := hlive.NewTag("a", "<h1>text_test</h1>")
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a>&lt;h1&gt;text_test&lt;/h1&gt;</a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagChildTag(t *testing.T) {
	t.Parallel()

	el := hlive.NewTag("a",
		hlive.NewTag("h1", "text_test"),
	)
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a><h1>text_test</h1></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagVoid(t *testing.T) {
	t.Parallel()

	el := hlive.T("hr", hlive.Attrs{"foo": "bar"})
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<hr foo="bar">`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestPage_RenderElementTagCSS(t *testing.T) {
	t.Parallel()

	el := hlive.T("hr",
		hlive.CSS{"c3": true},
		hlive.CSS{"c2": true},
		hlive.CSS{"c1": true},
		hlive.CSS{"c2": false})
	buff := bytes.NewBuffer(nil)

	if err := hlive.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<hr class="c3 c1">`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

// TODO: rethink now we have JavaScript
// func TestPage_Render(t *testing.T) {
// 	t.Parallel()
//
// 	page := hlive.NewPage(hlive.NewTag("title", "test_title"))
// 	buff := bytes.NewBuffer(nil)
//
// 	if err := page.HTML(buff); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	expect := `<!doctype html><html><head><meta charset="utf-8"><title>test_title</title></head><body></body></html>`
// 	if diff := deep.Equal(expect, buff.String()); diff != nil {
// 		t.Error(diff)
// 	}
// }
//
// func TestPage_RenderLang(t *testing.T) {
// 	t.Parallel()
//
// 	page := hlive.NewPage(nil)
// 	page.Lang = "en"
// 	buff := bytes.NewBuffer(nil)
//
// 	if err := page.HTML(buff); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	expect := `<!doctype html><html lang="en"><head><meta charset="utf-8"></head><body></body></html>`
// 	if diff := deep.Equal(expect, buff.String()); diff != nil {
// 		t.Error(diff)
// 	}
// }
//
// func TestPage_RenderBodyTag(t *testing.T) {
// 	t.Parallel()
//
// 	page := hlive.NewPage(nil)
// 	page.AddBody(hlive.NewTag("h1", hlive.CSS{"c1": true},
// 			"hi there",
// 		))
//
// 	buff := bytes.NewBuffer(nil)
// 	if err := page.HTML(buff); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	expect := `<!doctype html><html><head><meta charset="utf-8"></head><body><h1 class="c1">hi there</h1></body></html>`
// 	if diff := deep.Equal(expect, buff.String()); diff != nil {
// 		t.Error(diff)
// 	}
// }
