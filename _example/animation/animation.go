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

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		animationTarget := l.C("div", l.ClassBool{"animate__animated": true}, l.ClassBool{"text": true}, "HLive")

		animations := []string{
			"animate__hinge", "animate__jackInTheBox", "animate__rollIn", "animate__rollOut",
			"animate__bounce", "animate__flash", "animate__pulse", "animate__rubberBand", "animate__shakeX",
			"animate__shakeY", "animate__headShake", "animate__swing", "animate__tada", "animate__wobble",
			"animate__jello", "animate__heartBeat", "animate__flip", "animate__backInDown", "animate__backOutDown",
		}
		index := 0
		current := ""
		nextAnimation := func() {
			animationTarget.Add(l.ClassBool{animations[index]: false})

			index++
			if len(animations) == index {
				index = 0
			}

			current = animations[index]

			animationTarget.Add(l.ClassBool{animations[index]: true})
		}

		playing := false
		btnLabel := "Start"
		btn := l.C("button", &btnLabel,
			l.On("click", func(ctx context.Context, e l.Event) {
				if !playing {
					nextAnimation()
					btnLabel = "Stop"
				} else {
					btnLabel = "Start"
					current = ""
				}

				playing = !playing
			}),
			// You can create multiple event bindings for the same event and component
			l.On("click", func(ctx context.Context, e l.Event) {
				log.Println("INFO: Button Clicked")
			}),
		)

		animationTarget.Add(l.On("animationend", func(ctx context.Context, e l.Event) {
			if playing {
				nextAnimation()
			}
		}))
		animationTarget.Add(l.On("animationcancel", func(ctx context.Context, e l.Event) {
			playing = false
			btnLabel = "Start"
			current = ""
		}))

		page := l.NewPage()
		page.DOM().Title().Add("Animation Example")
		page.DOM().Head().Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))
		page.DOM().Head().Add(l.T("link",
			l.Attrs{"rel": "stylesheet", "href": "https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css"}))
		page.DOM().Head().Add(l.T("style", pageStyle))

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "CSS Animations"),
				l.T("p", "We can wait for the CSS animation to end before starting the next one"),
			),
			l.T("main",
				l.T("p", btn),
				l.T("p", "Current: ", &current),
				l.T("div", l.ClassBool{"box": true}, animationTarget),
			),
		)

		return page
	}

	return l.NewPageServer(f)
}
