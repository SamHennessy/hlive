package hlive

import (
	"context"
)

// Mounter wants to be notified after it's mounted.
type Mounter interface {
	UniqueTagger
	// Mount is called after a component is mounted
	Mount(ctx context.Context)
}

// Unmounter wants to be notified before it's unmounted.
type Unmounter interface {
	UniqueTagger
	// Unmount is called before a component is unmounted
	Unmount(ctx context.Context)
}

// Teardowner wants to have manual control when it needs to be removed from a Page.
// If you have a Mounter or Unmounter that will be permanently removed from a Page they must call the passed
// function to clean up their references.
type Teardowner interface {
	UniqueTagger
	// AddTeardown adds a teardown function
	AddTeardown(teardown func())
	// Teardown call the set teardown function passed in SetTeardown
	Teardown()
}

type ComponentMountable struct {
	*Component

	MountFunc   func(ctx context.Context)
	UnmountFunc func(ctx context.Context)
	teardowns   []func()
}

// CM is a shortcut for NewComponentMountable
func CM(name string, elements ...interface{}) *ComponentMountable {
	return NewComponentMountable(name, elements...)
}

func NewComponentMountable(name string, elements ...interface{}) *ComponentMountable {
	return &ComponentMountable{
		Component: NewComponent(name, elements...),
	}
}

func (c *ComponentMountable) Mount(ctx context.Context) {
	if c.MountFunc != nil {
		c.MountFunc(ctx)
	}
}

func (c *ComponentMountable) Unmount(ctx context.Context) {
	if c.UnmountFunc != nil {
		c.UnmountFunc(ctx)
	}
}

func (c *ComponentMountable) AddTeardown(teardown func()) {
	c.teardowns = append(c.teardowns, teardown)
}

func (c *ComponentMountable) Teardown() {
	for i := 0; i < len(c.teardowns); i++ {
		c.teardowns[i]()
	}
}

// WM is a shortcut for WrapMountable.
func WM(tag *Tag, elements ...interface{}) *ComponentMountable {
	return WrapMountable(tag, elements...)
}

// WrapMountable takes a Tag and creates a Component with it.
func WrapMountable(tag *Tag, elements ...interface{}) *ComponentMountable {
	return &ComponentMountable{
		Component: Wrap(tag, elements),
	}
}
