package hlive

import (
	"context"

	"github.com/rs/xid"
)

type Event struct {
	Binding  *EventBinding
	Value    string
	Key      string
	CharCode int
	KeyCode  int
	ShiftKey bool
	AltKey   bool
	CtrlKey  bool
}

type EventHandler func(ctx context.Context, e Event)

func NewEventBinding() *EventBinding {
	return &EventBinding{ID: xid.New().String()}
}

type EventBinding struct {
	ID        string
	Handler   EventHandler
	Type      EventType
	Component Componenter
}
