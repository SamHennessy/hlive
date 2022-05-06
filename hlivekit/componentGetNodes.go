package hlivekit

import (
	"github.com/SamHennessy/hlive"
)

// ComponentGetNodes add a custom GetNodes function to ComponentMountable
type ComponentGetNodes struct {
	*hlive.ComponentMountable

	GetNodesFunc func() *hlive.NodeGroup
}

// CGN is a shortcut for NewComponentGetNodes
func CGN(name string, getNodesFunc func() *hlive.NodeGroup, elements ...interface{}) *ComponentGetNodes {
	return NewComponentGetNodes(name, getNodesFunc, elements...)
}

func NewComponentGetNodes(name string, getNodesFunc func() *hlive.NodeGroup, elements ...interface{}) *ComponentGetNodes {
	return &ComponentGetNodes{
		ComponentMountable: hlive.NewComponentMountable(name, elements...),
		GetNodesFunc:       getNodesFunc,
	}
}

func (c *ComponentGetNodes) GetNodes() *hlive.NodeGroup {
	return c.GetNodesFunc()
}
