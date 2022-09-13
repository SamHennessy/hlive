package hlivekit

import (
	"errors"
	"fmt"

	l "github.com/SamHennessy/hlive"
)

var ErrInvalidListAdd = errors.New("value is not valid for a list")

// ComponentList is a way to manage a dynamic collection of Teardowner Node values. For example, the rows of a table.
//
// As the Node values in ComponentList are often added and removed there lies the possibility of memory leaks. To
// prevent this the items in the list must be Teardowner values. The list will call Teardown on each item as long as
// they are removed using its RemoveItem and RemoveAllItems functions.
//
// See NewComponentMountable, CM, WrapMountable, and WM for help with creating Teardowner values.
type ComponentList struct {
	*ComponentListSimple
}

// List is a shortcut for NewComponentList.
func List(name string, elements ...any) *ComponentList {
	return NewComponentList(name, elements...)
}

// NewComponentList returns a value of ComponentList
func NewComponentList(name string, elements ...any) *ComponentList {
	return &ComponentList{
		ComponentListSimple: NewComponentListSimple(name, elements...),
	}
}

// Add an element to this Component.
//
// You can add Groups, Teardowner, EventBinding, or None Node Elements
func (list *ComponentList) Add(elements ...any) {
	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		case *l.NodeGroup:
			g := v.Get()
			for j := 0; j < len(g); j++ {
				list.Add(g[j])
			}
		case l.Teardowner:
			list.AddItems(v)
		case *l.EventBinding:
			list.Component.Add(v)
		default:
			if l.IsNonNodeElement(v) {
				list.Tag.Add(v)
			} else {
				l.LoggerDev.Error().
					Str("callers", l.CallerStackStr()).
					Str("element", fmt.Sprintf("%#v", v)).
					Msg("invalid element")
			}
		}
	}
}

// AddItem allows you to add a node to the list
//
// Add nodes are often added and removed nodes needed to be a Teardowner.
// See NewComponentMountable, CM, WrapMountable, and WM for help with creating Teardowner values.
func (list *ComponentList) AddItem(items ...l.Teardowner) {
	for i := 0; i < len(items); i++ {
		list.ComponentListSimple.AddItems(items[i])
	}
}

// RemoveItems will remove a Teardowner can call its Teardown function.
func (list *ComponentList) RemoveItems(items ...l.Teardowner) {
	for i := 0; i < len(items); i++ {
		list.ComponentListSimple.RemoveItems(items[i])
		items[i].Teardown()
	}
}

// RemoveAllItems empties the list of items and calls Teardown on each of them.
func (list *ComponentList) RemoveAllItems() {
	for i := 0; i < len(list.ComponentListSimple.items); i++ {
		if td, ok := list.ComponentListSimple.items[i].(l.Teardowner); ok {
			td.Teardown()
		}
	}

	list.ComponentListSimple.RemoveAllItems()
}
