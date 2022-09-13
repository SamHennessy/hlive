package hlivekit

import (
	"fmt"
	"sync"

	l "github.com/SamHennessy/hlive"
)

// ComponentListSimple is a version of ComponentList that doesn't have the Teardown logic
type ComponentListSimple struct {
	*l.ComponentMountable

	items []l.UniqueTagger
	mu    sync.RWMutex
}

// NewComponentListSimple creates a ComponentListSimple value
func NewComponentListSimple(name string, elements ...any) *ComponentListSimple {
	list := &ComponentListSimple{
		ComponentMountable: l.CM(name),
	}

	list.Add(elements...)

	return list
}

// GetNodes returns the list of items.
func (list *ComponentListSimple) GetNodes() *l.NodeGroup {
	list.mu.RLock()
	defer list.mu.RUnlock()

	return l.G(list.items)
}

// Add an element to this ComponentListSimple.
//
// You can add Groups, UniqueTagger, EventBinding, or None Node Elements
func (list *ComponentListSimple) Add(elements ...any) {
	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		case *l.NodeGroup:
			list.Add(v.Get()...)
		case l.UniqueTagger:
			list.items = append(list.items, v)
		default:
			if l.IsNonNodeElement(v) {
				list.Component.Add(v)
			} else {
				l.LoggerDev.Warn().Str("callers", l.CallerStackStr()).
					Str("element", fmt.Sprintf("%#v", v)).
					Msg("invalid element type")
			}
		}
	}
}

func (list *ComponentListSimple) AddItems(items ...l.UniqueTagger) {
	list.mu.Lock()
	list.items = append(list.items, items...)
	list.mu.Unlock()
}

func (list *ComponentListSimple) RemoveItems(items ...l.UniqueTagger) {
	list.mu.Lock()
	defer list.mu.Unlock()

	var newList []l.UniqueTagger

	for i := 0; i < len(list.items); i++ {
		hit := false

		for j := 0; j < len(items); j++ {
			item := items[j]

			if item.GetID() == list.items[i].GetID() {
				hit = true

				break
			}
		}

		if !hit {
			newList = append(newList, list.items[i])
		}
	}

	list.items = newList
}

func (list *ComponentListSimple) RemoveAllItems() {
	list.mu.Lock()
	list.items = nil
	list.mu.Unlock()
}
