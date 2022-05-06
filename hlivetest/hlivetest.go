package hlivetest

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/playwright-community/playwright-go"
)

func FatalOnErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func Click(t *testing.T, pwpage playwright.Page, selector string) {
	t.Helper()

	FatalOnErr(t, pwpage.Click(selector))
}

func ClickAndWait(t *testing.T, pwpage playwright.Page, selector string) {
	t.Helper()

	done := AckWatcher(t, pwpage, selector)

	Click(t, pwpage, selector)

	<-done
}

func TextContent(t *testing.T, pwpage playwright.Page, selector string) string {
	t.Helper()

	text, err := pwpage.TextContent(selector)

	FatalOnErr(t, err)

	return text
}

func GetAttribute(t *testing.T, pwpage playwright.Page, selector string, attribute string) string {
	t.Helper()

	el, err := pwpage.QuerySelector(selector)
	FatalOnErr(t, err)

	val, err := el.GetAttribute(attribute)
	FatalOnErr(t, err)

	return val
}

func GetID(t *testing.T, pwpage playwright.Page, selector string) string {
	t.Helper()

	return GetAttribute(t, pwpage, selector, "id")
}

func Diff(t *testing.T, want interface{}, got interface{}) {
	t.Helper()

	if diff := deep.Equal(want, got); diff != nil {
		t.Error(diff)
	}
}

func DiffFatal(t *testing.T, want interface{}, got interface{}) {
	t.Helper()

	if diff := deep.Equal(want, got); diff != nil {
		t.Fatal(diff)
	}
}

func Title(t *testing.T, pwpage playwright.Page) string {
	t.Helper()

	title, err := pwpage.Title()
	FatalOnErr(t, err)

	return title
}
