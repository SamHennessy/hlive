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
	Component ComponentInterface
}
//
// func GetEventBindingFromTree(id string, children ...interface{}) *EventBinding {
// 	for i := 0; i < len(children); i++ {
// 		switch v := children[i].(type) {
// 		case []interface{}:
// 			{
// 				for j := 0; j < len(v); j++ {
// 					if a := GetEventBindingFromTree(id, v[j]); a != nil {
// 						return a
// 					}
// 				}
// 			}
// 		case ComponentInterface:
// 			if a := v.GetEventBinding(id); a != nil {
// 				return a
// 			}
//
// 			tagKids := v.Render()
// 			for j := 0; j < len(tagKids); j++ {
// 				if a := GetEventBindingFromTree(id, tagKids[j]); a != nil {
// 					return a
// 				}
// 			}
// 		case TagInterface:
// 			tagKids := v.Render()
// 			for j := 0; j < len(tagKids); j++ {
// 				if a := GetEventBindingFromTree(id, tagKids[j]); a != nil {
// 					return a
// 				}
// 			}
// 		}
// 	}
//
// 	return nil
// }
