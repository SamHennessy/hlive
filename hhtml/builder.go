package hhtml

import "github.com/SamHennessy/hlive"

func tagBuilder(name string, elements ...any) hlive.Adder {
	hasEventBinding := hasEventBinding(elements...)

	var el hlive.Adder
	if hasEventBinding {
		el = hlive.C(name, elements...)
	} else {
		el = hlive.T(name, elements...)
	}

	return el
}

func hasEventBinding(elements ...any) bool {
	for _, v := range elements {
		switch v := v.(type) {
		case *hlive.EventBinding:
			return true
		case []any:
			if hasEventBinding(v...) {
				return true
			}
		case *hlive.NodeGroup:
			if hasEventBinding(v.Get()...) {
				return true
			}
		case *hlive.ElementGroup:
			if hasEventBinding(v.Get()...) {
				return true
			}
		}
	}
	return false
}
