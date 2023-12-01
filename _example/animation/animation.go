package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

const pageStyle l.HTML = `
.box {
	overflow: hidden; 
	padding: 3em; 
	text-align: center;
	border: solid;
}
.text {
	display: inline-block; 
	font-size: 3em;
}
`

var animations = []string{
	"animate__hinge", "animate__jackInTheBox", "animate__rollIn", "animate__rollOut",
	"animate__bounce", "animate__flash", "animate__pulse", "animate__rubberBand", "animate__shakeX",
	"animate__shakeY", "animate__headShake", "animate__swing", "animate__tada", "animate__wobble",
	"animate__jello", "animate__heartBeat", "animate__flip", "animate__backInDown", "animate__backOutDown",
}

func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.Page {
	var (
		index    = l.Box(0)
		current  = l.Box("")
		playing  = l.NewLockBox(false)
		btnLabel = l.Box("Start")
	)

	animationTarget := l.C("div", l.Class("animate__animated text"), "HLive")

	nextAnimation := func() {
		index.Lock(func(v int) int {
			animationTarget.Add(l.ClassOff(animations[v]))

			v++
			if len(animations) == v {
				v = 0
			}

			current.Set(animations[v])
			animationTarget.Add(l.Class(animations[v]))

			return v
		})
	}

	animationTarget.Add(l.On("animationend", func(ctx context.Context, e l.Event) {
		if playing.Get() {
			nextAnimation()
		}
	}))

	animationTarget.Add(l.On("animationcancel", func(ctx context.Context, e l.Event) {
		playing.Set(false)
		btnLabel.Set("Start")
		current.Set("")
	}))

	btn := l.C("button", btnLabel,
		l.On("click", func(ctx context.Context, e l.Event) {
			playing.Lock(func(v bool) bool {
				if !v {
					nextAnimation()
					btnLabel.Set("Stop")
				} else {
					btnLabel.Set("Start")
					current.Set("")
				}

				return !v
			})
		}),
		// You can create multiple event bindings for the same event and component
		l.On("click", func(ctx context.Context, e l.Event) {
			log.Println("INFO: Button Clicked")
		}),
	)

	page := l.NewPage()
	page.DOM().Title().Add("Animation Example")
	page.DOM().Head().Add(
		l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}),
		l.T("link",
			l.Attrs{"rel": "stylesheet", "href": "https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css"}),
		l.T("style", pageStyle),
	)

	page.DOM().Body().Add(
		l.T("header",
			l.T("h1", "CSS Animations"),
			l.T("p", "We can wait for the CSS animation to end before starting the next one"),
		),
		l.T("main",
			l.T("p", btn),
			l.T("p", "Current: ", current),
			l.T("div", l.Class("box"), animationTarget),
		),
	)

	return page
}
