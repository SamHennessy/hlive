package hlivekit

import (
	"fmt"

	l "github.com/SamHennessy/hlive"
)

// ComponentListSimple is a version of ComponentList that doesn't have the Teardown logic
type ComponentListSimple struct {
	*l.ComponentMountable

	Items []l.UniqueTagger
}

// NewComponentListSimple creates a ComponentListSimple value
func NewComponentListSimple(name string, elements ...interface{}) *ComponentListSimple {
	list := &ComponentListSimple{
		ComponentMountable: l.CM(name),
	}

	list.Add(elements...)

	return list
}

// GetNodes returns the list of items.
func (list *ComponentListSimple) GetNodes() *l.NodeGroup {
	return l.G(list.Items)
}

// Add an element to this ComponentListSimple.
//
// You can add Groups, UniqueTagger, EventBinding, or None Node Elements
func (list *ComponentListSimple) Add(elements ...interface{}) {
	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		case *l.NodeGroup:
			g := v.Get()
			for j := 0; j < len(g); j++ {
				list.Add(g[j])
			}
		case l.UniqueTagger:
			list.AddItems(v)
		case *l.EventBinding:
			list.Component.Add(v)
		default:
			if l.IsNonNodeElement(v) {
				list.Tag.Add(v)
			} else {
				panic(fmt.Errorf("ComponentListSimple.Add: element: %#v: %w", v, ErrInvalidListAdd))
			}
		}
	}
}

func (list *ComponentListSimple) AddItems(items ...l.UniqueTagger) {
	for i := 0; i < len(items); i++ {
		if !l.IsNode(items[i]) {
			panic(fmt.Sprintf("component list: passed item is not a node: %v", items[i]))
		}
	}

	for i := 0; i < len(items); i++ {
		list.Items = append(list.Items, items[i])
	}
}

func (list *ComponentListSimple) RemoveItems(items ...l.UniqueTagger) {
	var newList []l.UniqueTagger

	for i := 0; i < len(list.Items); i++ {
		hit := false

		for j := 0; j < len(items); j++ {
			item := items[j]

			if item.GetID() == list.Items[i].GetID() {
				hit = true

				break
			}
		}

		if !hit {
			newList = append(newList, list.Items[i])
		}
	}

	list.Items = newList
}

func (list *ComponentListSimple) RemoveAllItems() {
	list.Items = nil
}
