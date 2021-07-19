package hlive

import (
	"context"

	"github.com/rs/xid"
)

type Event struct {
	// The binding that was listening for this event
	Binding *EventBinding
	// If an input has a value set by the browsers on page load, different to the inputs value attribute this type of
	// event is sent. This typically happens on page reload after data has been inputted to a field.
	IsInitial bool
	// The value of the field, if relevant
	Value string
	// Used when an event source could have multiple values
	Values []string
	// Selected is true, for the element interacted with, if a radio or checkbox is checked or a select option is selected.
	// Most relevant for checkbox as it always has a value, this lets you know if they are currently checked or not.
	Selected bool
	// Key related values are only used on keyboard related events
	Key      string
	CharCode int
	KeyCode  int
	ShiftKey bool
	AltKey   bool
	CtrlKey  bool
	// Used for file inputs and uploads
	File *File
}

type File struct {
	// File name
	Name string
	// Size of the file in bytes
	Size int
	// Mime type
	Type string
	// The file contents
	Data []byte
	// Which file is this in the total file count, 0 index
	Index int
	// How many files are being uploaded in total
	Total int
}

type EventHandler func(ctx context.Context, e Event)

func NewEventBinding() *EventBinding {
	return &EventBinding{ID: xid.New().String()}
}

type EventBinding struct {
	// Unique ID for this binding
	ID string
	// Function to call when binding is triggered
	Handler EventHandler
	// Component we are bound to
	Component Componenter
	// Call this binding once then discard it
	Once bool
	// Name of the JavaScript event that will trigger this binding
	Name string
}
