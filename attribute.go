package hlive

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type Attributer interface {
	GetName() string
	GetValue() string
	GetValuePtr() *string
	SetValue(string)
	SetValuePtr(*string)
	IsNoEscapeString() bool
	Clone() *Attribute
}

type AttributePluginer interface {
	Attributer

	// Initialize will only be called once per attribute name for diff render
	Initialize(page *Page)
	// InitializeSSR will only be called once per attribute name for server side render
	InitializeSSR(page *Page)
}

// Attrs is a helper for adding Attributes to nodes
// You can update an existing Attribute by adding new Attrs, it's also possible to pass a string by reference.
// You can remove an Attribute by passing a nil value.
type Attrs map[string]any

func (a Attrs) GetAttributers() []Attributer {
	newAttrs := make([]Attributer, 0, len(a))

	for name, val := range a {
		attr := NewAttributePtr(name, nil)
		switch v := val.(type) {
		case string:
			attr.SetValue(v)
		case *string:
			attr.SetValuePtr(v)
		case nil:
			// Nop
		default:
			LoggerDev.Warn().Str("value", fmt.Sprintf("%#v", val)).
				Msg("Only string, *string, and nil are valid for Attr values")
		}

		newAttrs = append(newAttrs, attr)
	}

	return newAttrs
}

// ClassBool a special Attribute for working with CSS classes on nodes using a bool to toggle them on and off.
// It supports turning them on and off and allowing overriding. Due to how Go maps work the order of the classes in
// the map is not preserved.
// All Classes are de-duped, overriding a Class by adding new ClassBool will result in the old Class getting updated.
// You don't have to use ClassBool to add a class attribute, but it's the recommended way to do it.
type ClassBool map[string]bool

// TODO: add tests and docs
type (
	Class        string
	ClassOff     string
	ClassList    []string
	ClassListOff []string
)

// Style is a special Attribute that allows you to work with ClassBool styles on nodes.
// It allows you to override
// All Styles are de-duped, overriding a Style by adding new Style will result in the old Style getting updated.
// You don't have to use Style to add a style attribute, but it's the recommended way to do it.
type Style map[string]any

// NewAttribute create a new Attribute
func NewAttribute(name string, value string) *Attribute {
	return NewAttributePtr(name, &value)
}

// NewAttribute create a new Attribute
func NewAttributePtr(name string, value *string) *Attribute {
	return &Attribute{name: strings.ToLower(name), value: value}
}

// Attribute represents an HTML attribute e.g. id="submitBtn"
type Attribute struct {
	// name must always be lowercase
	name           string
	value          *string
	noEscapeString bool
	mu             sync.RWMutex
}

func (a *Attribute) MarshalMsgpack() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return msgpack.Marshal([2]any{a.name, a.value})
}

func (a *Attribute) UnmarshalMsgpack(b []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var values [2]any
	if err := msgpack.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("msgpack.Unmarshal: %w", err)
	}

	a.name, _ = values[0].(string)

	val, _ := values[1].(string)
	a.value = &val

	return nil
}

func (a *Attribute) GetName() string {
	if a == nil {
		return ""
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.name
}

func (a *Attribute) SetValue(value string) {
	a.value = &value
}

func (a *Attribute) GetValue() string {
	if a == nil || a.value == nil {
		return ""
	}

	return *a.value
}

func (a *Attribute) SetValuePtr(value *string) {
	a.value = value
}

func (a *Attribute) GetValuePtr() *string {
	if a == nil || a.value == nil {
		return nil
	}

	return a.value
}

// Clone creates a new Attribute using the data from this Attribute
func (a *Attribute) Clone() *Attribute {
	a.mu.Lock()
	defer a.mu.Unlock()
	newA := NewAttributePtr(a.name, nil)
	newA.noEscapeString = a.noEscapeString

	// Copy the value only
	if a.GetValuePtr() != nil {
		val := a.GetValue()
		newA.value = &val
	}

	return newA
}

func (a *Attribute) IsNoEscapeString() bool {
	if a == nil {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.noEscapeString
}

func (a *Attribute) SetNoEscapeString(noEscapeString bool) {
	a.mu.Lock()
	a.noEscapeString = noEscapeString
	a.mu.Unlock()
}

func (a *Attribute) GetAttribute() *Attribute {
	return a
}

func anyToAttributes(attrs ...any) []Attributer {
	var newAttrs []Attributer

	for i := 0; i < len(attrs); i++ {
		if attrs[i] == nil {
			continue
		}

		switch v := attrs[i].(type) {
		case Attrs:
			newAttrs = append(newAttrs, v.GetAttributers()...)
		case Attributer:
			newAttrs = append(newAttrs, v)
		case []Attributer:
			newAttrs = append(newAttrs, v...)
		case []*Attribute:
			for j := 0; j < len(v); j++ {
				newAttrs = append(newAttrs, v[j])
			}
		default:
			LoggerDev.Error().
				Str("callers", CallerStackStr()).
				Str("value", fmt.Sprintf("%#v", v)).
				Msg("invalid attribute")

			continue
		}
	}

	return newAttrs
}
