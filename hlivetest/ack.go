package hlivetest

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/playwright-community/playwright-go"
	"github.com/teris-io/shortid"
)

//go:embed ack.js
var ackJavaScript []byte

const (
	ackAttrName   = "data-hlive-test-ack"
	ackIDAttrName = "data-hlive-test-ack-id"
	ackExtraKey   = "test-ack-id"
	ackCtxKey     = "browser_testing.ack"
)

func Ack() l.Attributer {
	return &ack{
		l.NewAttribute(ackAttrName, ""),
	}
}

type ack struct {
	*l.Attribute
}

func (a *ack) Initialize(page *l.Page) {
	page.HookBeforeEventAdd(ackBeforeEvent)
	page.HookAfterRenderAdd(ackAfterRender)
	page.DOM().Head().Add(l.T("script", l.HTML(ackJavaScript)))
}

// Look in the extra data for ack id and add the value to the context if found
func ackBeforeEvent(ctx context.Context, e l.Event) (context.Context, l.Event) {
	if e.Extra[ackExtraKey] != "" {
		return context.WithValue(ctx, ackCtxKey, e.Extra[ackExtraKey]), e
	}

	return ctx, e
}

// If ack id in context then send a message
func ackAfterRender(ctx context.Context, diffs []l.Diff, send chan<- l.MessageWS) {
	ackID, ok := ctx.Value(ackCtxKey).(string)
	if !ok || ackID == "" {
		return
	}

	send <- l.MessageWS{Message: []byte("ack|" + ackID + "\n")}
}

// TODO: detect timeout and errors
func AckWatcher(t *testing.T, page playwright.Page, selector string) <-chan error {
	t.Helper()

	id := shortid.MustGenerate()

	_, err := page.EvalOnSelector(selector, "node => node.setAttribute(\"data-hlive-test-ack-id\", \""+id+"\")")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)

	go func() {
		if _, err := page.WaitForFunction("hliveTestAck.received[\""+id+"\"] === true", nil); err != nil {
			done <- fmt.Errorf("wait for function: %w", err)

			return
		}

		done <- nil
	}()

	return done
}
