package hlivekit

import (
	"context"
	"io"
	"sync"

	"github.com/SamHennessy/hlive"
	"github.com/teris-io/shortid"
)

type QueueMessage struct {
	Topic string
	Value interface{}
}

type QueueSubscriber interface {
	GetID() string
	OnMessage(message QueueMessage)
}

type PubSubMounter interface {
	GetID() string
	PubSubMount(context.Context, *PubSub)
}

type PubSubSSRMounter interface {
	GetID() string
	PubSubSSRMount(context.Context, *PubSub)
}

type PubSub struct {
	subsLock    sync.Mutex
	subscribers map[string][]QueueSubscriber
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: map[string][]QueueSubscriber{},
	}
}

func (ps *PubSub) Subscribe(sub QueueSubscriber, topics ...string) {
	if len(topics) == 0 {
		hlive.LoggerDev.Warn().Str("callers", hlive.CallerStackStr()).Msg("no topics passed")

		return
	}

	if sub == nil {
		hlive.LoggerDev.Warn().Str("callers", hlive.CallerStackStr()).Msg("sub nil")

		return
	}

	ps.subsLock.Lock()
	defer ps.subsLock.Unlock()

	for i := 0; i < len(topics); i++ {
		ps.subscribers[topics[i]] = append(ps.subscribers[topics[i]], sub)
	}
}

func (ps *PubSub) SubscribeFunc(subFunc func(message QueueMessage), topics ...string) SubscribeFunc {
	sub := NewSub(subFunc)

	ps.Subscribe(sub, topics...)

	return sub
}

func (ps *PubSub) Unsubscribe(sub QueueSubscriber, topics ...string) {
	if len(topics) == 0 {
		hlive.LoggerDev.Warn().Str("callers", hlive.CallerStackStr()).Msg("no topics passed")
	}

	if sub == nil {
		hlive.LoggerDev.Warn().Str("callers", hlive.CallerStackStr()).Msg("sub when nil")

		return
	}

	ps.subsLock.Lock()
	defer ps.subsLock.Unlock()

	for i := 0; i < len(topics); i++ {
		var newList []QueueSubscriber

		for j := 0; j < len(ps.subscribers[topics[i]]); j++ {
			if ps.subscribers[topics[i]][j].GetID() == sub.GetID() {
				continue
			}

			newList = append(newList, ps.subscribers[topics[i]][j])
		}

		ps.subscribers[topics[i]] = newList
	}
}

func (ps *PubSub) Publish(topic string, value any) {
	item := QueueMessage{topic, value}
	for i := 0; i < len(ps.subscribers[topic]); i++ {
		ps.subscribers[topic][i].OnMessage(item)
	}
}

type SubscribeFunc struct {
	fn func(message QueueMessage)
	id string
}

func (s SubscribeFunc) OnMessage(message QueueMessage) {
	s.fn(message)
}

func (s SubscribeFunc) GetID() string {
	return s.id
}

func NewSub(onMessageFn func(message QueueMessage)) SubscribeFunc {
	return SubscribeFunc{onMessageFn, shortid.MustGenerate()}
}

const PipelineProcessorKeyPubSubMount = "hlivekit_ps_mount"

func (a *PubSubAttribute) PipelineProcessorPubSub() *hlive.PipelineProcessor {
	pp := hlive.NewPipelineProcessor(PipelineProcessorKeyPubSubMount)

	pp.BeforeTagger = func(ctx context.Context, w io.Writer, tag hlive.Tagger) (hlive.Tagger, error) {
		if comp, ok := tag.(PubSubMounter); ok {
			a.lock.Lock()

			if _, exists := a.mountedMap[comp.GetID()]; !exists {
				a.mountedMap[comp.GetID()] = struct{}{}

				comp.PubSubMount(ctx, a.pubSub)
			}

			// A way to remove the key when you delete a Component
			if comp, ok := tag.(hlive.Teardowner); ok {
				comp.AddTeardown(func() {
					a.lock.Lock()
					delete(a.mountedMap, comp.GetID())
					a.lock.Unlock()
				})
			}

			a.lock.Unlock()
		}

		return tag, nil
	}

	return pp
}

const PubSubAttributeName = "data-hlive-pubsub"

func InstallPubSub(pubSub *PubSub) hlive.Attributer {
	attr := &PubSubAttribute{
		Attribute:  hlive.NewAttribute(PubSubAttributeName, ""),
		pubSub:     pubSub,
		mountedMap: map[string]struct{}{},
	}

	return attr
}

type PubSubAttribute struct {
	*hlive.Attribute

	pubSub *PubSub
	// Will memory leak is you don't use a Teardowner when deleting Components
	mountedMap map[string]struct{}
	lock       sync.Mutex
	rendered   bool
}

func (a *PubSubAttribute) Initialize(page *hlive.Page) {
	if a.rendered {
		return
	}

	page.PipelineDiff.Add(a.PipelineProcessorPubSub())
}

func (a *PubSubAttribute) InitializeSSR(page *hlive.Page) {
	a.rendered = true
	page.PipelineDiff.Add(a.PipelineProcessorPubSub())
}

// ComponentPubSub add PubSub to ComponentMountable
type ComponentPubSub struct {
	*hlive.ComponentMountable

	MountPubSubFunc func(ctx context.Context, pubSub *PubSub)
}

// CPS is a shortcut for NewComponentPubSub
func CPS(name string, elements ...interface{}) *ComponentPubSub {
	return NewComponentPubSub(name, elements...)
}

func NewComponentPubSub(name string, elements ...interface{}) *ComponentPubSub {
	return &ComponentPubSub{
		ComponentMountable: hlive.NewComponentMountable(name, elements...),
	}
}

func (c *ComponentPubSub) PubSubMount(ctx context.Context, pubSub *PubSub) {
	if c.MountPubSubFunc == nil {
		return
	}

	c.MountPubSubFunc(ctx, pubSub)
}
