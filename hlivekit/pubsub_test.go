package hlivekit_test

import (
	"sync"
	"testing"

	"github.com/SamHennessy/hlive/hlivekit"
	"github.com/go-test/deep"
	"github.com/teris-io/shortid"
)

type testSubscriber struct {
	id          string
	called      bool
	calledTopic string
	calledValue interface{}
	wait        sync.WaitGroup
}

func newSub() *testSubscriber {
	sub := &testSubscriber{id: shortid.MustGenerate()}
	sub.wait.Add(1)

	return sub
}

func (s *testSubscriber) OnMessage(message hlivekit.QueueMessage) {
	s.called = true
	s.calledTopic = message.Topic
	s.calledValue = message.Value
	s.wait.Done()
}

func (s *testSubscriber) GetID() string {
	return s.id
}

func TestNewPubSub(t *testing.T) {
	t.Parallel()

	if hlivekit.NewPubSub() == nil {
		t.Error("nil returned")
	}
}

func TestPubSub_PublishAndSubscribe(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1")

	ps.Publish("topic_2", nil)

	if sub.called {
		t.Fatal("unexpected sub call")
	}

	ps.Publish("topic_1", "foo")

	if !sub.called {
		t.Fatal("expected sub call")
	}

	if diff := deep.Equal("foo", sub.calledValue); diff != nil {
		t.Error(diff)
	}
}

func TestPubSub_PublishAsync(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1")

	go ps.Publish("topic_1", nil)

	sub.wait.Wait()

	if !sub.called {
		t.Fatal("expected sub call")
	}
}

func TestPubSub_SubscribeMultiTopic(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1", "topic_2")

	ps.Publish("topic_1", nil)

	if !sub.called || sub.calledTopic != "topic_1" {
		t.Fatal("expected sub call")
	}

	sub.called = false
	sub.wait.Add(1)

	ps.Publish("topic_2", nil)

	if !sub.called || sub.calledTopic != "topic_2" {
		t.Fatal("expected sub call")
	}
}

func TestPubSub_Unsubscribe(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1")

	ps.Unsubscribe(sub, "topic_1")

	ps.Publish("topic_1", nil)

	if sub.called {
		t.Fatal("unexpected sub call")
	}
}

func TestPubSub_UnsubscribeOneOfMulti(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1", "topic_2")

	ps.Unsubscribe(sub, "topic_1")

	ps.Publish("topic_1", nil)

	if sub.called {
		t.Fatal("unexpected sub call")
	}

	ps.Publish("topic_2", nil)

	if !sub.called {
		t.Fatal("expected sub call")
	}
}

func TestPubSub_UnsubscribeMulti(t *testing.T) {
	t.Parallel()

	sub := newSub()
	ps := hlivekit.NewPubSub()

	ps.Subscribe(sub, "topic_1", "topic_2")

	ps.Unsubscribe(sub, "topic_1", "topic_2")

	ps.Publish("topic_1", nil)

	if sub.called {
		t.Fatal("unexpected sub call")
	}

	ps.Publish("topic_2", nil)

	if sub.called {
		t.Fatal("unexpected sub call")
	}
}
