package hlive_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestEvent_OnHandler(t *testing.T) {
	t.Parallel()

	var val bool

	eb := l.On("input", func(_ context.Context, _ l.Event) {
		val = true
	})

	if val {
		t.Fatal("handler call before expected")
	}

	eb.Handler(nil, l.Event{})

	if !val {
		t.Error("unexpected handler")
	}
}

func TestEvent_On(t *testing.T) {
	t.Parallel()

	eb := l.On("INPUT", nil)

	if eb.Once {
		t.Error("once not default to false")
	}

	if diff := deep.Equal("input", eb.Name); diff != nil {
		t.Error(diff)
	}
}

func TestEvent_OnOnce(t *testing.T) {
	t.Parallel()

	if eb := l.OnOnce("input", nil); !eb.Once {
		t.Error("once not default to true")
	}
}
