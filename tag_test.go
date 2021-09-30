package hlive_test

import (
	"bytes"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestTag_T(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	div := l.T("div")
	hr := l.T("hr")

	if diff := deep.Equal(false, div.IsVoid()); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(true, hr.IsVoid()); diff != nil {
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
