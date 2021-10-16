package hlive_test

import (
	"bytes"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/go-test/deep"
)

func TestRenderer_CSS(t *testing.T) {
	t.Parallel()

	el := l.T("hr",
		l.CSS{"c3": true},
		l.CSS{"c2": true},
		l.CSS{"c1": true},
		l.CSS{"c2": false})
	buff := bytes.NewBuffer(nil)

	if err := l.NewRender().HTML(buff, el); err != nil {
		t.Fatal(err)
	}

	css := el.GetAttribute("class")
	if css == nil {
		t.Fatal("attribute not found")
	}

	if diff := deep.Equal("c3 c1", css.GetAttribute().GetValue()); diff != nil {
		t.Error(diff)
	}
}
