package hlive

import (
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type Attributer interface {
	GetName() string
	GetValue() string
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

// Attrs is a helper for adding and updating Attributes to nodes
type Attrs map[string]string

func (a Attrs) GetAttributers() []Attributer {
	newAttrs := make([]Attributer, 0, len(a))

	for name, val := range a {
		newAttrs = append(newAttrs, NewAttribute(name, val))
	}

	return newAttrs
}

type AttrsLockBox map[string]*LockBox[string]

func (a AttrsLockBox) GetAttributers() []Attributer {
	newAttrs := make([]Attributer, 0, len(a))

	for name, val := range a {
		newAttrs = append(newAttrs, NewAttributeLockBox(name, val))
	}

	return newAttrs
}

// AttrsOff a helper for removing Attributes
type AttrsOff []string

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

// Style is a special Element that allows you to work the properties of style attribute.
// A property and value will be added or updated.
// You don't have to use Style to add a style attribute, but it's the recommended way to do it.
type Style map[string]string

// StyleLockBox like Style but, you can update the property values indirectly
// TODO: add test
type StyleLockBox map[string]*LockBox[string]

// StyleOff remove an existing style property, ignored if the property doesn't exist
// TODO: add test
type StyleOff []string

// NewAttribute create a new Attribute
func NewAttribute(name string, value string) *Attribute {
	return &Attribute{name: strings.ToLower(name), value: NewLockBox(value)}
}

// NewAttributeLockBox create a new Attribute using the passed LockBox value
func NewAttributeLockBox(name string, value *LockBox[string]) *Attribute {
	return &Attribute{name: strings.ToLower(name), value: value}
}

// Attribute represents an HTML attribute e.g. id="submitBtn"
type Attribute struct {
	// name must always be lowercase
	name           string
	value          *LockBox[string]
	noEscapeString bool
}

func (a *Attribute) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal([2]any{a.name, a.value.Get()})
}

func (a *Attribute) UnmarshalMsgpack(b []byte) error {
	var values [2]any
	if err := msgpack.Unmarshal(b, &values); err != nil {
		return fmt.Errorf("msgpack.Unmarshal: %w", err)
	}

	a.name, _ = values[0].(string)

	val, _ := values[1].(string)
	a.value = NewLockBox(val)

	return nil
}

func (a *Attribute) GetName() string {
	if a == nil {
		return ""
	}

	return a.name
}

func (a *Attribute) SetValue(value string) {
	a.value.Set(value)
}

func (a *Attribute) GetValue() string {
	if a == nil {
		return ""
	}

	return a.value.Get()
}

// Clone creates a new Attribute using the data from this Attribute
func (a *Attribute) Clone() *Attribute {
	newA := NewAttribute(a.name, a.value.Get())
	newA.noEscapeString = a.noEscapeString

	return newA
}

func (a *Attribute) IsNoEscapeString() bool {
	if a == nil {
		return false
	}

	return a.noEscapeString
}

func (a *Attribute) SetNoEscapeString(noEscapeString bool) {
	a.noEscapeString = noEscapeString
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
		case AttrsLockBox:
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
