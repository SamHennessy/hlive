package main

import (
	"context"
	"net/http"
	"os"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	http.Handle("/", Home(logger))

	logger.Info().Str("addr", ":3000").Msg("listing")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func Home(logger zerolog.Logger) *l.PageServer {
	f := func() *l.Page {
		h := l.C("div", l.CSS{"animate__animated": true}, l.CSS{"text": true}, "HLive")

		animations := []string{"animate__hinge", "animate__jackInTheBox", "animate__rollIn", "animate__rollOut", "animate__bounce", "animate__flash", "animate__pulse", "animate__rubberBand", "animate__shakeX", "animate__shakeY", "animate__headShake", "animate__swing", "animate__tada", "animate__wobble", "animate__jello", "animate__heartBeat", "animate__flip", "animate__backInDown", "animate__backOutDown"}
		index := 0
		current := ""
		nextAnimation := func() {
			h.Add(l.CSS{animations[index]: false})

			index++
			if len(animations) == index {
				index = 0
			}

			current = animations[index]

			h.Add(l.CSS{animations[index]: true})
		}

		playing := false
		btnLabel := "Start"
		btn := l.C("button", &btnLabel,
			l.OnClick(func(ctx context.Context, e l.Event) {
				if !playing {
					nextAnimation()
					btnLabel = "Stop"
				} else {
					btnLabel = "Start"
					current = ""
				}

				playing = !playing
			}),
			l.OnClick(func(ctx context.Context, e l.Event) {
				logger.Info().Msg("Button Clicked")
			}),
		)

		h.On(l.OnAnimationEnd(func(ctx context.Context, e l.Event) {
			if playing {
				nextAnimation()
			}
		}))
		h.On(l.OnAnimationCancel(func(ctx context.Context, e l.Event) {
			playing = false
			btnLabel = "Start"
			current = ""
		}))

		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Animation Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css"}))
		page.Head.Add(l.T("style",
			l.HTML(`
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
`),
		))
		page.Body.Add(
			btn,
			&current,
			l.T("hr"),
			l.T("div", l.CSS{"box": true}, h),
		)

		return page
	}

	return l.NewPageServer(f)
}
