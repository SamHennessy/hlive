package hlive_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
)

func TestPage_CloseHooks(t *testing.T) {
	t.Parallel()

	p := l.NewPage()

	var called bool

	hook := func(ctx context.Context, page *l.Page) {
		called = true
	}

	p.HookCloseAdd(hook)

	p.Close(context.Background())

	if !called {
		t.Error("close hook not called")
	}
}
