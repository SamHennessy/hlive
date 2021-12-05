package hlive_test

import (
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

func TestTag_AddNodeTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  interface{}
	}{
		{"nil", nil},
		{"string", "string"},
		{"*string", new(string)},
		{"HTML", l.HTML("")},
		{"*HTML", new(l.HTML)},
		{"Tagger", &testTagger{}},
		{"UniqueTagger", &testUniqueTagger{}},
		{"Componenter", &testComponenter{}},
		{"[]interface{}", []interface{}{}},
		{"NodeGroup", l.G()},
		{"[]*Tag", []*l.Tag{}},
		{"[]Tagger", []l.Tagger{}},
		{"[]*Component", []*l.Component{}},
		{"[]Componenter", []l.Componenter{}},
		{"[]UniqueTagger", []l.UniqueTagger{}},
		{"float34", float32(1)},
		{"float64", float64(1)},
		{"*float34", new(float32)},
		{"*float64", new(float64)},
		{"int", 1},
		{"int8", int8(1)},
		{"int16", int16(1)},
		{"int32", int32(1)},
		{"int64", int64(1)},
		{"*int", new(int)},
		{"*int8", new(int8)},
		{"*int16", new(int16)},
		{"*int32", new(int32)},
		{"*int64", new(int64)},
		{"uint", uint(1)},
		{"uint8", uint8(1)},
		{"uint16", uint16(1)},
		{"uint32", uint32(1)},
		{"uint64", uint64(1)},
		{"*uint", new(uint)},
		{"*uint8", new(uint8)},
		{"*uint16", new(uint16)},
		{"*uint32", new(uint32)},
		{"*uint64", new(uint64)},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tag := l.T("div")
			tag.Add(tt.arg)
		})
	}
}

func TestTag_AddElementTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  interface{}
	}{
		{"*Attribute", l.NewAttribute("value")},
		{"[]*Attribute", []*l.Attribute{}},
		{"Attrs", l.Attrs{}},
		{"ClassBool", l.ClassBool{}},
		{"Style", l.Style{}},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tag := l.T("div")
			tag.Add(tt.arg)
		})
	}
}

func TestTag_AddNode(t *testing.T) {
	t.Parallel()

	parent := l.T("div")
	parent.Add(l.T("div"))

	if len(parent.GetNodes().Get()) != 1 {
		t.Fatalf("expected 1 child got %v", len(parent.GetNodes().Get()))
	}

	parent.Add("foo")

	if len(parent.GetNodes().Get()) != 2 {
		t.Fatalf("expected 2 children got %v", len(parent.GetNodes().Get()))
	}
}

func TestTag_AddNodes(t *testing.T) {
	t.Parallel()

	parent := l.T("div")
	parent.Add(l.T("div"), "foo")

	if len(parent.GetNodes().Get()) != 2 {
		t.Fatal("expected 2 children")
	}
}
