package hlive_test

import (
	"testing"

	"github.com/SamHennessy/hlive"
)

func TestIsValidElement(t *testing.T) {
	t.Parallel()

	type args struct {
		el any
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"bool", args{true}, false},
		{"nil", args{nil}, true},
		{"string", args{"test"}, true},
		{"html", args{hlive.HTML("<h1>title</h1>")}, true},
		{"tag", args{hlive.T("h1")}, true},
		{"css", args{hlive.ClassBool{"c1": true}}, true},
		{"attribute", args{hlive.AttrsOff{"disabled"}}, true},
		{"attrs", args{hlive.Attrs{"href": "https://foo.com"}}, true},
		{"component", args{hlive.C("span")}, true},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hlive.IsElement(tt.args.el); got != tt.want {
				t.Errorf("IsElement() = %v, want %v", got, tt.want)
			}
		})
	}
}
