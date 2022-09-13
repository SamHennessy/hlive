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

	mountFunc   func(ctx context.Context)
	unmountFunc func(ctx context.Context)
	teardowns   []func()
}

// CM is a shortcut for NewComponentMountable
func CM(name string, elements ...any) *ComponentMountable {
	return NewComponentMountable(name, elements...)
}

func NewComponentMountable(name string, elements ...any) *ComponentMountable {
	return &ComponentMountable{
		Component: NewComponent(name, elements...),
	}
}

func (c *ComponentMountable) Mount(ctx context.Context) {
	if c == nil {
		return
	}

	c.Tag.mu.RLock()
	f := c.mountFunc
	c.Tag.mu.RUnlock()

	if f != nil {
		f(ctx)
	}
}

func (c *ComponentMountable) Unmount(ctx context.Context) {
	if c == nil {
		return
	}

	c.Tag.mu.RLock()
	f := c.unmountFunc
	c.Tag.mu.RUnlock()

	if c.unmountFunc != nil {
		f(ctx)
	}
}

func (c *ComponentMountable) SetMount(mount func(ctx context.Context)) {
	c.Tag.mu.Lock()
	c.mountFunc = mount
	c.Tag.mu.Unlock()
}

func (c *ComponentMountable) SetUnmount(unmount func(ctx context.Context)) {
	c.Tag.mu.Lock()
	c.unmountFunc = unmount
	c.Tag.mu.Unlock()
}

func (c *ComponentMountable) AddTeardown(teardown func()) {
	c.Tag.mu.Lock()
	c.teardowns = append(c.teardowns, teardown)
	c.Tag.mu.Unlock()
}

func (c *ComponentMountable) Teardown() {
	c.Tag.mu.RLock()
	teardowns := c.teardowns
	c.Tag.mu.RUnlock()

	for i := 0; i < len(teardowns); i++ {
		teardowns[i]()
	}
}

// WM is a shortcut for WrapMountable.
func WM(tag *Tag, elements ...any) *ComponentMountable {
	return WrapMountable(tag, elements...)
}

// WrapMountable takes a Tag and creates a Component with it.
func WrapMountable(tag *Tag, elements ...any) *ComponentMountable {
	return &ComponentMountable{
		Component: Wrap(tag, elements),
	}
}
