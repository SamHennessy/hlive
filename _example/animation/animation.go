package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

const pageStyle l.HTML = `
.box {
	overflow: hidden; 
	background-color: aliceblue; 
	padding: 3em; 
	text-align: center;
}
.text {
	display: inline-block; 
	font-size: 3em;
}
`

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/", home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		animationTarget := l.C("div", l.CSS{"animate__animated": true}, l.CSS{"text": true}, "HLive")

		animations := []string{
			"animate__hinge", "animate__jackInTheBox", "animate__rollIn", "animate__rollOut",
			"animate__bounce", "animate__flash", "animate__pulse", "animate__rubberBand", "animate__shakeX",
			"animate__shakeY", "animate__headShake", "animate__swing", "animate__tada", "animate__wobble",
			"animate__jello", "animate__heartBeat", "animate__flip", "animate__backInDown", "animate__backOutDown",
		}
		index := 0
		current := ""
		nextAnimation := func() {
			animationTarget.Add(l.CSS{animations[index]: false})

			index++
			if len(animations) == index {
				index = 0
			}

			current = animations[index]

			animationTarget.Add(l.CSS{animations[index]: true})
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
				logger.Info().Msg("Button Clicked")
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
		page.SetLogger(logger)
		page.Title.Add("Animation Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))
		page.Head.Add(l.T("link",
			l.Attrs{"rel": "stylesheet", "href": "https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css"}))
		page.Head.Add(l.T("style", pageStyle))

		page.Body.Add(
			btn,
			&current,
			l.T("div", l.CSS{"box": true}, animationTarget),
		)

		return page
	}

	return l.NewPageServer(f)
}
