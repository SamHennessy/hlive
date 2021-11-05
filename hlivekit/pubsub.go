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
	OnMessage(item QueueMessage)
}

type PubSubMounter interface {
	GetID() string
	PubSubMount(*PubSub)
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
	ps.subsLock.Lock()
	defer ps.subsLock.Unlock()

	for i := 0; i < len(topics); i++ {
		ps.subscribers[topics[i]] = append(ps.subscribers[topics[i]], sub)
	}
}

func (ps *PubSub) Unsubscribe(sub QueueSubscriber, topics ...string) {
	ps.subsLock.Lock()
	defer ps.subsLock.Unlock()

	for i := 0; i < len(topics); i++ {
		var newList []QueueSubscriber

		for j := 0; j < len(ps.subscribers[topics[i]]); j++ {
			if ps.subscribers[topics[i]][j] == sub {
				continue
			}

			newList = append(newList, ps.subscribers[topics[i]][j])
		}

		ps.subscribers[topics[i]] = newList
	}
}

func (ps *PubSub) Publish(topic string, value interface{}) {
	item := QueueMessage{topic, value}
	for i := 0; i < len(ps.subscribers[topic]); i++ {
		ps.subscribers[topic][i].OnMessage(item)
	}
}

type subFN struct {
	fn func(message QueueMessage)
	id string
}

func (s subFN) OnMessage(item QueueMessage) {
	s.fn(item)
}

func (s subFN) GetID() string {
	return s.id
}

func NewSub(onMessageFn func(message QueueMessage)) QueueSubscriber {
	return subFN{onMessageFn, shortid.MustGenerate()}
}

func PipelineProcessorPubSub(pubSub *PubSub) *hlive.PipelineProcessor {
	pp := hlive.NewPipelineProcessor(hlive.PipelineProcessorKeyPubSubMount)

	// Will memory leak is you don't use a Teardowner when deleting Components
	cache := map[string]struct{}{}

	pp.BeforeTagger = func(ctx context.Context, w io.Writer, tag hlive.Tagger) (hlive.Tagger, error) {
		if comp, ok := tag.(PubSubMounter); ok {
			if _, exists := cache[comp.GetID()]; !exists {
				cache[comp.GetID()] = struct{}{}

				comp.PubSubMount(pubSub)
			}

			// A way to remove the key when you delete a Component
			if comp, ok := tag.(hlive.Teardowner); ok {
				comp.AddTeardown(func() {
					delete(cache, comp.GetID())
				})
			}
		}

		return tag, nil
	}

	return pp
}

const PubSubAttributeName = "data-hlive-pubsub"

func InstallPubSub(pubSub *PubSub) hlive.Attributer {
	attr := &PubSubAttribute{
		Attribute: hlive.NewAttribute(PubSubAttributeName, ""),
		pubSub:    pubSub,
	}

	return attr
}

type PubSubAttribute struct {
	*hlive.Attribute

	pubSub *PubSub
}

func (a *PubSubAttribute) Initialize(page *hlive.Page) {
	page.PipelineDiff.AddBefore(hlive.PipelineProcessorKeyEventBindingCache, PipelineProcessorPubSub(a.pubSub))
	page.PipelineRender.AddBefore(hlive.PipelineProcessorKeyEventBindingCache, PipelineProcessorPubSub(a.pubSub))
}
