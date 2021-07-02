package hlive

type ComponentList struct {
	*Component

	Items []Componenter
}

func NewComponentList(tagName string) *ComponentList {
	return &ComponentList{
		Component: NewComponent(tagName),
	}
}

func (list *ComponentList) GetNodes() []interface{} {
	return Tree(list.Items)
}

func (list *ComponentList) AddItem(items ...Componenter) {
	list.Items = append(list.Items, items...)
}

func (list *ComponentList) RemoveItem(items ...Componenter) {
	var newList []Componenter

	for i := 0; i < len(list.Items); i++ {
		for j := 0; j < len(items); j++ {
			item := items[i]

			if item.GetID() == list.Items[i].GetID() {
				continue
			}
		}

		newList = append(newList, list.Items[i])
	}

	list.Items = newList
}

func (list *ComponentList) RemoveAllItems() {
	list.Items = nil
}
