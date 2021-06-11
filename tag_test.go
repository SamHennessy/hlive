package hlive_test

import (
	"bytes"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestTag_T(t *testing.T) {
	tag := l.T("div")

	if tag == nil {
		t.Fatal("nil returned")
	}

	if diff := deep.Equal("div", tag.GetName()); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(0, len(tag.GetAttributes())); diff != nil {
		t.Error(diff)
	}
}

func TestTag_IsVoid(t *testing.T) {
	div := l.T("div")
	hr  := l.T("hr")

	if diff := deep.Equal(false, div.IsVoid()); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(true, hr.IsVoid()); diff != nil {
		t.Error(diff)
	}
}

func TestTag_Attributes(t *testing.T) {
	div := l.T("div")

	a := l.NewAttribute("foo", "bar")
	b := l.NewAttribute("biz", "baz")

	div.SetAttributes(a, b)

	if diff := deep.Equal([]*l.Attribute{a,b}, div.GetAttributes()); diff != nil {
		t.Error(diff)
	}

	div.RemoveAttributes("foo")

	if diff := deep.Equal([]*l.Attribute{b}, div.GetAttributes()); diff != nil {
		t.Error(diff)
	}
}
func TestTag_Attrs(t *testing.T) {
	div := l.T("div", l.Attrs{"foo":"bar"}, l.Attrs{"biz":"baz"})

	a := l.NewAttribute("foo", "bar")
	b := l.NewAttribute("biz", "baz")

	attrsResult := div.GetAttributes()
	if diff := deep.Equal(a.Value, attrsResult[0].Value); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(b.Value, attrsResult[1].Value); diff != nil {
		t.Error(diff)
	}
}

func TestTag_CSS(t *testing.T) {
	t.Parallel()

	el := l.T("hr",
		l.CSS{"c3": true},
		l.CSS{"c2": true},
		l.CSS{"c1": true},
		l.CSS{"c2": false})
	buff := bytes.NewBuffer(nil)

	if err := l.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	css := el.GetAttribute("class")
	if css == nil {
		t.Fatal("attribute not found")
	}

	if diff := deep.Equal("c3 c1", css.GetValue()); diff != nil {
		t.Error(diff)
	}
}

func TestTag_Style(t *testing.T) {
	t.Parallel()

	tag := l.T("div", l.Style{"display": "none"})

	style := tag.GetAttribute("style")
	if style == nil {
		t.Fatal("attribute not found")
	}

	if diff := deep.Equal("display:none;", style.GetValue()); diff != nil {
		t.Error(diff)
	}
}

func TestTag_StyleMulti(t *testing.T) {
	t.Parallel()

	tag := l.T("div",
		l.Style{"display": "none"},
		l.Style{"padding": "3em"},
		l.Style{"text-align": "center"},
	)
	tag.Add(l.Style{"padding": nil})

	style := tag.GetAttribute("style")
	if style == nil {
		t.Fatal("attribute not found")
	}

	if diff := deep.Equal("display:none;text-align:center;", style.GetValue()); diff != nil {
		t.Error(diff)
	}
}
