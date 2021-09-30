package hlive

import "fmt"

type ComponentList struct {
	*ComponentMountable

	Items []UniqueTagger
}

func NewComponentList(name string, elements ...interface{}) *ComponentList {
	return &ComponentList{
		ComponentMountable: CM(name, elements...),
	}
}

func (list *ComponentList) GetNodes() interface{} {
	return list.Items
}

func (list *ComponentList) AddItems(items ...UniqueTagger) {
	for i := 0; i < len(items); i++ {
		if !IsNode(items[i]) {
			panic(fmt.Sprintf("component list: passed item is not a node: %v", items[i]))
		}
	}

	for i := 0; i < len(items); i++ {
		list.Items = append(list.Items, items[i])
	}
}

func (list *ComponentList) RemoveItems(items ...UniqueTagger) {
	var newList []UniqueTagger

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

func (list *ComponentList) RemoveAllItems() {
	list.Items = nil
}

type ComponentListTidy struct {
	*ComponentList
}

// List is a shortcut for NewComponentListTidy.
func List(name string, elements ...interface{}) *ComponentListTidy {
	return NewComponentListTidy(name, elements...)
}

func NewComponentListTidy(name string, elements ...interface{}) *ComponentListTidy {
	return &ComponentListTidy{
		ComponentList: NewComponentList(name, elements...),
	}
}

func (list *ComponentListTidy) AddItem(items ...Teardowner) {
	for i := 0; i < len(items); i++ {
		list.ComponentList.AddItems(items[i])
	}
}

func (list *ComponentListTidy) RemoveItem(items ...Teardowner) {
	for i := 0; i < len(items); i++ {
		list.ComponentList.RemoveItems(items[i])
		items[i].Teardown()
	}
}

func (list *ComponentListTidy) RemoveAllItems() {
	for i := 0; i < len(list.ComponentList.Items); i++ {
		if td, ok := list.ComponentList.Items[i].(Teardowner); ok {
			td.Teardown()
		}
	}

	list.ComponentList.RemoveAllItems()
}
