package hlive_test

import (
	"strings"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestAttribute_GetValueSetValue(t *testing.T) {
	t.Parallel()

	a := l.NewAttributePtr("foo", nil)

	if diff := deep.Equal("", a.GetValue()); diff != nil {
		t.Error(diff)
	}

	a.SetValue("bar")

	if diff := deep.Equal("bar", a.GetValue()); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_SetValueByReference(t *testing.T) {
	t.Parallel()

	a := l.NewAttributePtr("foo", nil)

	attrVal := "bar"

	a.SetValuePtr(&attrVal)

	if diff := deep.Equal("bar", a.GetValue()); diff != nil {
		t.Error(diff)
	}

	attrVal = "fizz"

	if diff := deep.Equal("fizz", a.GetValue()); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagAddAttributesGetAttributes(t *testing.T) {
	t.Parallel()

	div := l.T("div")

	a := l.NewAttribute("foo", "bar")
	b := l.NewAttribute("biz", "baz")

	div.AddAttributes(a, b)

	if diff := deep.Equal([]l.Attributer{a, b}, div.GetAttributes()); diff != nil {
		t.Error(diff)
	}

	div.RemoveAttributes("foo")

	if diff := deep.Equal([]l.Attributer{b}, div.GetAttributes()); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagAttrs(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Attrs{"foo": "bar"}, l.Attrs{"biz": "baz"})

	a := l.NewAttribute("foo", "bar")
	b := l.NewAttribute("biz", "baz")

	if diff := deep.Equal(a.GetValue(), div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(b.GetValue(), div.GetAttributeValue("biz")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagRemoveAttribute(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.NewAttribute("foo", "bar"))

	if diff := deep.Equal(1, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}

	div.RemoveAttributes("foo")

	if diff := deep.Equal(0, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagNewAttributeRemove(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.NewAttribute("foo", "bar"))

	if diff := deep.Equal(1, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}

	div.Add(l.NewAttributePtr("foo", nil))

	if diff := deep.Equal(0, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagAttrsRemove(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Attrs{"foo": "bar"})

	if diff := deep.Equal(1, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}

	div.Add(l.Attrs{"foo": nil})

	if diff := deep.Equal(0, len(div.GetAttributes())); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagAttrsReferenceValue(t *testing.T) {
	t.Parallel()

	attrVal := "bar"

	div := l.T("div", l.Attrs{"foo": &attrVal})

	if diff := deep.Equal("bar", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}

	attrVal = "baz"

	if diff := deep.Equal("baz", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagAttrsUpdateValue(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Attrs{"foo": "bar"})

	if diff := deep.Equal("bar", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.Attrs{"foo": "baz"})

	if diff := deep.Equal("baz", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_TagNewAttributeUpdateValue(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.NewAttribute("foo", "bar"))

	if diff := deep.Equal("bar", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.NewAttribute("foo", "baz"))

	if diff := deep.Equal("baz", div.GetAttributeValue("foo")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_CSS(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.ClassBool{"foo": true})

	if diff := deep.Equal("foo", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}

	div.Add("div", l.ClassBool{"bar": true})

	if diff := deep.Equal("foo bar", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}

	div.Add("div", l.ClassBool{"foo": false})

	if diff := deep.Equal("bar", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}

	div.Add("div", l.ClassBool{"foo": true})

	if diff := deep.Equal("bar foo", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_CSSMultiUnordered(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.ClassBool{"foo": true, "bar": true, "fizz": true})

	if !strings.Contains(div.GetAttributeValue("class"), "foo") {
		t.Error("foo not found")
	}

	if !strings.Contains(div.GetAttributeValue("class"), "bar") {
		t.Error("bar not found")
	}

	if !strings.Contains(div.GetAttributeValue("class"), "fizz") {
		t.Error("fizz not found")
	}

	div.Add(l.ClassBool{"bar": false})

	if strings.Contains(div.GetAttributeValue("class"), "bar") {
		t.Error("bar found")
	}
}

func TestAttribute_CSSMultiOrdered(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.ClassBool{"foo": true}, l.ClassBool{"bar": true}, l.ClassBool{"fizz": true})

	if diff := deep.Equal("foo bar fizz", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.ClassBool{"bar": false})

	if diff := deep.Equal("foo fizz", div.GetAttributeValue("class")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_Style(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Style{"foo": "bar"})

	if diff := deep.Equal("foo:bar;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}

	div.Add("div", l.Style{"foo": nil})

	if diff := deep.Equal("", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_StyleOverride(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Style{"foo": "bar"})

	if diff := deep.Equal("foo:bar;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}

	div.Add("div", l.Style{"foo": "fizz"})

	if diff := deep.Equal("foo:fizz;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_StyleMultiUnordered(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Style{"foo": "a", "bar": "b", "fizz": "c"})

	if !strings.Contains(div.GetAttributeValue("style"), "foo:a;") {
		t.Error("foo:a; not found")
	}

	if !strings.Contains(div.GetAttributeValue("style"), "bar:b;") {
		t.Error("bar:a; not found")
	}

	if !strings.Contains(div.GetAttributeValue("style"), "fizz:c;") {
		t.Error("fizz:c; not found")
	}

	div.Add(l.Style{"bar": nil})

	if strings.Contains(div.GetAttributeValue("style"), "bar:b;") {
		t.Error("bar:a; found")
	}
}

func TestAttribute_StyleMultiOrdered(t *testing.T) {
	t.Parallel()

	div := l.T("div", l.Style{"foo": "a"}, l.Style{"bar": "b"}, l.Style{"fizz": "c"})

	if diff := deep.Equal("foo:a;bar:b;fizz:c;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.Style{"bar": "z"})

	if diff := deep.Equal("foo:a;bar:z;fizz:c;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.Style{"bar": nil})

	if diff := deep.Equal("foo:a;fizz:c;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}

	div.Add(l.Style{"bar": "x"})

	if diff := deep.Equal("foo:a;fizz:c;bar:x;", div.GetAttributeValue("style")); diff != nil {
		t.Error(diff)
	}
}

func TestAttribute_Clone(t *testing.T) {
	t.Parallel()

	a := l.NewAttribute("foo", "bar")
	b := a.Clone()

	if diff := deep.Equal(a, b); diff != nil {
		t.Error(diff)
	}

	if a == b {
		t.Error("attributes are the same")
	}
}
