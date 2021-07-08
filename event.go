package hlive

import (
	"context"

	"github.com/rs/xid"
)

type Event struct {
	Binding   *EventBinding
	IsInitial bool
	Value     string
	Key       string
	CharCode  int
	KeyCode   int
	ShiftKey  bool
	AltKey    bool
	CtrlKey   bool
	File      *File
}

type File struct {
	Name string
	Size int
	Type string
	Data []byte
	// Which file is this in the total file count, 0 index
	Index int
	// How many files are being uploaded in total
	Total int
}

func (f *File) GetData() []byte {
	return nil
}

type EventHandler func(ctx context.Context, e Event)

func NewEventBinding() *EventBinding {
	return &EventBinding{ID: xid.New().String()}
}

type EventBinding struct {
	ID        string
	Handler   EventHandler
	Component Componenter
	Once      bool

	Name string
}
