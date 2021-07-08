package main

import (
	"context"
	"fmt"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	s := newService()

	http.HandleFunc("/", sessionMiddleware(home(logger, s).ServeHTTP))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger, s *service) *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("HTTP Session Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("h1", "Your Message"),
			l.T("p", "Your message will persist between page reloads but not server reloads"),
			newMessage(s),
		)

		return page
	}

	return l.NewPageServer(f)
}

type ctxKey string

const SessionKey ctxKey = "session"

func sessionMiddleware(h http.HandlerFunc) http.HandlerFunc {
	cookieName := "hlive_session"

	return func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		cook, err := r.Cookie(cookieName)
		switch {
		case err == http.ErrNoCookie:
			sessionID = xid.New().String()
			http.SetCookie(w,
				&http.Cookie{Name: cookieName, Value: sessionID, Path: "/", SameSite: http.SameSiteStrictMode})
		case err != nil:
			fmt.Println("ERROR: get cookie: ", err.Error())
		default:
			sessionID = cook.Value
		}

		r = r.WithContext(context.WithValue(r.Context(), SessionKey, sessionID))

		h(w, r)
	}
}

func GetSessionID(ctx context.Context) string {
	val, _ := ctx.Value(SessionKey).(string)

	return val
}

func newService() *service {
	return &service{userMessage: map[string]string{}}
}

type service struct {
	userMessage map[string]string
}

func (s *service) SetMessage(userID, message string) {
	s.userMessage[userID] = message
}

func (s *service) GetMessage(userID string) string {
	return s.userMessage[userID]
}

func newMessage(service *service) *message {
	c := &message{
		Component: l.C("textarea"),
		service:   service,
	}

	c.Add(l.On("change", func(ctx context.Context, e l.Event) {
		c.service.SetMessage(GetSessionID(ctx), e.Value)
	}))

	return c
}

type message struct {
	*l.Component

	Message string

	service *service
}

func (c *message) Mount(ctx context.Context) {
	c.Message = c.service.GetMessage(GetSessionID(ctx))
}

func (c *message) GetNodes() interface{} {
	return l.Tree(c.Message)
}
