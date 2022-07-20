package hlive_test

import (
	"bytes"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestRenderer_RenderElementTagCSS(t *testing.T) {
	t.Parallel()

	el := l.T("hr",
		l.ClassBool{"c3": true},
		l.ClassBool{"c2": true},
		l.ClassBool{"c1": true},
		l.ClassBool{"c2": false})
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<hr class="c3 c1">`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementText(t *testing.T) {
	t.Parallel()

	buff := bytes.NewBuffer(nil)
	if err := l.NewRenderer().HTML(buff, "<h1>text_test</h1>"); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("&lt;h1&gt;text_test&lt;/h1&gt;", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementRawHTML(t *testing.T) {
	t.Parallel()

	el := l.HTML("<h1>html_test</h1>")
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("<h1>html_test</h1>", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTag(t *testing.T) {
	t.Parallel()

	el := l.NewTag("a")
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("<a></a>", buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTagAttr(t *testing.T) {
	t.Parallel()

	el := l.NewTag("a", l.NewAttribute("href", "https://example.com"))
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a href="https://example.com"></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTagAttrs(t *testing.T) {
	t.Parallel()

	el := l.NewTag("a", l.Attrs{"href": "https://example.com"})
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a href="https://example.com"></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTagChildText(t *testing.T) {
	t.Parallel()

	el := l.NewTag("a", "<h1>text_test</h1>")
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a>&lt;h1&gt;text_test&lt;/h1&gt;</a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTagChildTag(t *testing.T) {
	t.Parallel()

	el := l.NewTag("a",
		l.NewTag("h1", "text_test"),
	)
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<a><h1>text_test</h1></a>`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_RenderElementTagVoid(t *testing.T) {
	t.Parallel()

	el := l.T("hr", l.Attrs{"foo": "bar"})
	buff := bytes.NewBuffer(nil)

	if err := l.NewRenderer().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(`<hr foo="bar">`, buff.String()); diff != nil {
		t.Error(diff)
	}
}

func TestRenderer_Attribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attrs   l.Attrs
		wantW   string
		wantErr bool
	}{
		{
			"simple",
			l.Attrs{"foo": "bar"},
			` foo="bar"`,
			false,
		},
		{
			"empty",
			l.Attrs{"foo": ""},
			` foo=""`,
			false,
		},
		{
			"json",
			l.Attrs{"foo": `["key1"]`},
			` foo="[&#34;key1&#34;]"`,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := l.NewRenderer()
			w := &bytes.Buffer{}
			if err := r.Attribute(tt.attrs.GetAttributes(), w); (err != nil) != tt.wantErr {
				t.Errorf("Attribute() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Attribute() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
