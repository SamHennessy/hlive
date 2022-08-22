package hlive_test

import (
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestComponent_GetID(t *testing.T) {
	t.Parallel()

	c := l.C("div")
	c.SetID("1")
	b := l.C("div")
	b.SetID("2")

	if c.GetID() == "" || b.GetID() == "" {
		t.Error("id is an empty string")
	}

	if c.GetID() == b.GetID() {
		t.Error("id not unique")
	}

	if diff := deep.Equal(c.GetID(), c.GetAttributeValue(l.AttrID)); diff != nil {
		t.Error(diff)
	}
}

func TestComponent_IsAutoRender(t *testing.T) {
	t.Parallel()

	c := l.C("div")

	if !c.IsAutoRender() {
		t.Error("auto render not true by default")
	}

	c.AutoRender = false

	if c.IsAutoRender() {
		t.Error("not able to set auto render")
	}
}

func TestComponent_AddAttribute(t *testing.T) {
	t.Parallel()

	c := l.C("div")
	c.SetID("1")

	eb1 := l.On("input", nil)
	eb2 := l.On("click", nil)

	if c.GetAttributeValue(l.AttrOn) != "" {
		t.Errorf("unexpected value for %s = %s", l.AttrOn, c.GetAttributeValue(l.AttrOn))
	}

	c.Add(eb1)

	expected := eb1.ID + "|" + eb1.Name
	if diff := deep.Equal(expected, c.GetAttributeValue(l.AttrOn)); diff != nil {
		t.Error(diff)
	}

	c.Add(eb2)

	expected = eb1.ID + "|" + eb1.Name + "," + eb2.ID + "|" + eb2.Name
	if diff := deep.Equal(expected, c.GetAttributeValue(l.AttrOn)); diff != nil {
		t.Error(diff)
	}
}

func TestComponent_AddGetEventBinding(t *testing.T) {
	t.Parallel()

	eb1 := l.On("input", nil)

	c := l.C("div", eb1)

	if c.GetEventBinding(eb1.ID) == nil {
		t.Error("event binding not found")
	}
}

func TestComponent_AddRemoveEventBinding(t *testing.T) {
	t.Parallel()

	eb1 := l.On("input", nil)
	eb2 := l.On("click", nil)

	c := l.C("div", eb1, eb2)
	c.SetID("1")

	c.RemoveEventBinding(eb1.ID)

	if c.GetEventBinding(eb1.ID) != nil {
		t.Error("event binding not removed")
	}

	if c.GetEventBinding(eb2.ID) == nil {
		t.Error("event binding not found")
	}
}

func TestComponent_Wrap(t *testing.T) {
	t.Parallel()

	tag := l.T("div")
	c := l.W(tag)

	if c.Tag != tag {
		t.Error("tag not wrapped")
	}
}
