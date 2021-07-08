package hlive_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SamHennessy/hlive"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-test/deep"
	"github.com/rs/zerolog"
)

type fixtures struct {
	ctx     context.Context
	cleanUp func(t *testing.T)
	ts      *httptest.Server
}

func setup(t *testing.T, body []interface{}) *fixtures {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
		// chromedp.WithDebugf(log.Printf),
		chromedp.WithErrorf(log.Printf),
	)

	home := func() *hlive.PageServer {
		f := func() *hlive.Page {
			page := hlive.NewPage()
			page.SetLogger(zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.WarnLevel))
			page.Body.Add(body)

			return page
		}

		return hlive.NewPageServer(f)
	}

	f := &fixtures{
		ctx: ctx,
		ts:  httptest.NewServer(home()),
	}

	f.cleanUp = func(t *testing.T) {
		if chromedp.FromContext(ctx).Browser != nil {
			if err := chromedp.Cancel(ctx); err != nil {
				t.Error("cancel browser:", err)
			}
		}

		f.ts.Close()
		cancel()
	}

	return f
}

func respOK(t *testing.T, err error, resp *network.Response) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(int64(http.StatusOK), resp.Status); diff != nil {
		t.Error(diff)
	}
}

func errFatal(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func navWait(t *testing.T, ctx context.Context, url, selector string) {
	t.Helper()

	resp, err := chromedp.RunResponse(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector),
	)
	respOK(t, err, resp)
}

func getText(t *testing.T, ctx context.Context, selector string) string {
	t.Helper()

	var content string

	errFatal(t, chromedp.Run(ctx, chromedp.Text(selector, &content)))

	return content
}

func TestBrowser_ClickButton(t *testing.T) {
	t.Parallel()

	count := 0
	cBtn := hlive.C("button", hlive.Attrs{"id": "target"},
		hlive.On("click", func(ctx context.Context, e hlive.Event) {
			count++
		}),
		&count,
	)

	f := setup(t, hlive.Tree(cBtn))
	defer f.cleanUp(t)

	// var nodes []*cdp.Node
	var content string
	resp, err := chromedp.RunResponse(f.ctx,
		chromedp.Navigate(f.ts.URL),
		// chromedp.Nodes(`document`, &nodes, chromedp.ByJSPath),
		chromedp.WaitVisible(`#target`),
		chromedp.Text("#target", &content),
	)

	respOK(t, err, resp)

	if diff := deep.Equal("0", content); diff != nil {
		t.Error(diff)
	}

	// fmt.Println("Content:", content)
	// fmt.Println("Document tree:")
	// fmt.Println(nodes[0].Dump("  ", "  ", false))

	// click
	err = chromedp.Run(f.ctx,
		chromedp.Click("#target"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// read
	err = chromedp.Run(f.ctx,
		// chromedp.Nodes(`document`, &nodes, chromedp.ByJSPath),
		chromedp.Text("#target", &content),
	)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal("1", content); diff != nil {
		t.Error(diff)
	}

	// fmt.Println("Content:", content)
	// fmt.Println("Document tree:")
	// fmt.Println(nodes[0].Dump("  ", "  ", false))

	// t.Fail()
}

func TestBrowser_ClickButtonMultiple(t *testing.T) {
	t.Parallel()

	count := 0
	cBtn := hlive.C("button", hlive.Attrs{"id": "target"},
		hlive.On("click", func(ctx context.Context, e hlive.Event) {
			count++
		}),
		&count,
	)

	f := setup(t, hlive.Tree(cBtn))
	defer f.cleanUp(t)

	navWait(t, f.ctx, f.ts.URL, "#target")

	if diff := deep.Equal("0", getText(t, f.ctx, "#target")); diff != nil {
		t.Error(diff)
	}

	// click
	for i := 0; i < 5; i++ {
		errFatal(t, chromedp.Run(f.ctx, chromedp.Click("#target")))
	}

	if diff := deep.Equal("5", getText(t, f.ctx, "#target")); diff != nil {
		t.Error(diff)
	}
}
